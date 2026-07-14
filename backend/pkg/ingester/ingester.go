package ingester

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"ragpack/pkg/chunker"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/meta"
	"ragpack/pkg/telemetry"
)

const (
	batchSize    = 100
	pollInterval = 30 * time.Second
)

type Ingester interface {
	Start(ctx context.Context, workers int)
	Submit(job meta.Job, r io.ReadCloser)
	Stop()
}

type queueItem struct {
	job    meta.Job
	reader io.ReadCloser // non-nil for direct uploads, nil for URI-based
	etag   *string       // set only by the auto-refresh scheduler, when it found a new ETag
}

type WorkerPool struct {
	queue     chan queueItem
	inFlight  sync.Map // jobID -> struct{}: currently enqueued (in the channel) or being processed by this process
	metaStore meta.MetaStore
	vectorDb  db.VectorDb
	registry  *embedder.Registry
	chunkCfg  chunker.Config
	limiter   *rate.Limiter
	telemetry *telemetry.Recorder
	waitGroup sync.WaitGroup
}

func New(metaStore meta.MetaStore, vectorDb db.VectorDb, registry *embedder.Registry, workers int, ratePerSec float64, chunkCfg chunker.Config, rec *telemetry.Recorder) Ingester {
	return &WorkerPool{
		queue:     make(chan queueItem, workers*10),
		metaStore: metaStore,
		vectorDb:  vectorDb,
		registry:  registry,
		chunkCfg:  chunkCfg,
		limiter:   rate.NewLimiter(rate.Limit(ratePerSec), 1),
		telemetry: rec,
	}
}

func (wp *WorkerPool) Start(ctx context.Context, workers int) {
	go wp.loop(ctx, workers)
	go wp.refreshScheduler(ctx)
}

func (wp *WorkerPool) Submit(job meta.Job, r io.ReadCloser) {
	wp.enqueue(queueItem{job: job, reader: r})
}

// enqueue skips a job already in-flight (inFlight.LoadOrStore is the atomic arbiter, not DB status) and otherwise blocks until it's placed on the channel.
func (wp *WorkerPool) enqueue(item queueItem) {
	if _, dup := wp.inFlight.LoadOrStore(item.job.ID, struct{}{}); dup {
		return
	}
	wp.queue <- item
}

// tryEnqueue is enqueue for callers that can't block on a full channel — un-marks the job on failure so a later attempt can retry it.
func (wp *WorkerPool) tryEnqueue(item queueItem) bool {
	if _, dup := wp.inFlight.LoadOrStore(item.job.ID, struct{}{}); dup {
		return false
	}
	select {
	case wp.queue <- item:
		return true
	default:
		wp.inFlight.Delete(item.job.ID)
		return false
	}
}

func (wp *WorkerPool) Stop() {
	wp.waitGroup.Wait()
}

func (wp *WorkerPool) loop(ctx context.Context, workers int) {
	// re-queue pending and processing (processing = crashed mid-job at last run)
	wp.requeue(ctx, true)

	for i := 0; i < workers; i++ {
		wp.waitGroup.Add(1)
		go wp.run(ctx)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// only re-queue pending on periodic poll — processing means a worker has it
			wp.requeue(ctx, false)
		}
	}
}

func (wp *WorkerPool) requeue(ctx context.Context, includeProcessing bool) {
	statuses := []meta.JobStatus{meta.JobStatusPending}
	if includeProcessing {
		statuses = append(statuses, meta.JobStatusProcessing)
	}

	var jobs []meta.Job
	for _, status := range statuses {
		s := status
		js, err := wp.metaStore.ListJobs(ctx, meta.JobFilter{Status: &s}, 10_000, 0)
		if err != nil {
			log.Printf("ingester: requeue list %s jobs: %v", status, err)
			continue
		}
		jobs = append(jobs, js...)
	}

	requeued := 0
	for _, j := range jobs {
		if wp.tryEnqueue(queueItem{job: j}) {
			requeued++
		}
	}

	if requeued > 0 {
		log.Printf("ingester: re-queued %d jobs", requeued)
	}
}

func (wp *WorkerPool) run(ctx context.Context) {
	defer wp.waitGroup.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-wp.queue:
			if err := wp.processJob(ctx, item); err != nil {
				log.Printf("ingester: job %s failed: %v", item.job.ID, err)
			}
			wp.inFlight.Delete(item.job.ID)
		}
	}
}
