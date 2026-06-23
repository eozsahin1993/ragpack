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
}

type WorkerPool struct {
	queue    chan queueItem
	metaStore meta.MetaStore
	vectorDb  db.VectorDb
	registry  *embedder.Registry
	chunkCfg  chunker.Config
	limiter   *rate.Limiter
	wg        sync.WaitGroup
}

func New(metaStore meta.MetaStore, vectorDb db.VectorDb, registry *embedder.Registry, workers int, ratePerSec float64) Ingester {
	return &WorkerPool{
		queue:     make(chan queueItem, workers*10),
		metaStore: metaStore,
		vectorDb:  vectorDb,
		registry:  registry,
		chunkCfg:  chunker.DefaultConfig(),
		limiter:   rate.NewLimiter(rate.Limit(ratePerSec), 1),
	}
}

func (wp *WorkerPool) Start(ctx context.Context, workers int) {
	go wp.loop(ctx, workers)
}

func (wp *WorkerPool) Submit(job meta.Job, r io.ReadCloser) {
	wp.queue <- queueItem{job: job, reader: r}
}

func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
}

func (wp *WorkerPool) loop(ctx context.Context, workers int) {
	// re-queue pending and processing (processing = crashed mid-job at last run)
	wp.requeue(ctx, true)

	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
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
		js, err := wp.metaStore.ListJobsByStatus(ctx, status)
		if err != nil {
			log.Printf("ingester: requeue list %s jobs: %v", status, err)
			continue
		}
		jobs = append(jobs, js...)
	}

	for _, j := range jobs {
		select {
		case wp.queue <- queueItem{job: j}:
		default:
			// queue full — job stays in SQLite and will be picked up next poll
		}
	}

	if len(jobs) > 0 {
		log.Printf("ingester: re-queued %d jobs", len(jobs))
	}
}

func (wp *WorkerPool) run(ctx context.Context) {
	defer wp.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-wp.queue:
			if err := wp.processJob(ctx, item); err != nil {
				log.Printf("ingester: job %s failed: %v", item.job.ID, err)
			}
		}
	}
}
