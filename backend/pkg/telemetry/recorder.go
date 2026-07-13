// Package telemetry records RAG pipeline events (ingestion, query, RAG) into
// local Parquet files for the admin analytics dashboard. Recording is
// fire-and-forget: producers fill a typed event struct and hand it to the
// Recorder, which buffers and flushes in a background goroutine. Nothing here
// is exposed on the public API surface. Each event type lives in its own
// subpackage (ingestion, query, trace) — see sink.go for how Recorder stays
// generic over them.
package telemetry

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"ragpack/pkg/telemetry/ingestion"
	"ragpack/pkg/telemetry/query"
	"ragpack/pkg/telemetry/trace"
)

const (
	flushInterval = 60 * time.Second
	flushRows     = 500
	queueSize     = 4096
)

type Config struct {
	Enabled       bool
	Dir           string
	RetentionDays int
	MaxSizeMB     int
	RedactText    bool
}

// Recorder buffers events and flushes them to Parquet in the background.
// A nil *Recorder is valid and drops everything — callers never nil-check.
type Recorder struct {
	cfg       Config
	ch        chan any
	sinks     []sink
	closed    chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
	dropped   atomic.Int64
}

func New(cfg Config) *Recorder {
	if !cfg.Enabled {
		return nil
	}
	r := &Recorder{
		cfg:    cfg,
		ch:     make(chan any, queueSize),
		closed: make(chan struct{}),
		sinks: []sink{
			newSink(ingestion.Table),
			newSink(query.Table),
			newSink(trace.Table),
		},
	}
	r.wg.Add(2)
	go r.loop()
	go r.janitor()
	return r
}

// recordable is implemented by every event type via a pointer receiver (see
// each package's Prepare). Prepare stamps server-generated fields in place
// and reports whether the event should be recorded at all — redaction and
// drop rules genuinely differ per type (a redacted query event still
// records with QueryText blanked; a redacted or ID-less trace is dropped
// entirely), so that decision stays with the event type, not here.
type recordable interface {
	Prepare(redact bool) bool
}

func (r *Recorder) Record(ev recordable) {
	if r == nil {
		return
	}
	if !ev.Prepare(r.cfg.RedactText) {
		return
	}
	r.enqueue(ev)
}

// enqueue never blocks: the pipeline must not wait on telemetry, so a full
// queue drops the event instead.
func (r *Recorder) enqueue(e any) {
	if r == nil {
		return
	}
	select {
	case <-r.closed:
	case r.ch <- e:
	default:
		if n := r.dropped.Add(1); n == 1 || n%1000 == 0 {
			log.Printf("telemetry: queue full, %d events dropped so far", n)
		}
	}
}

// Close flushes buffered events and stops the background goroutines.
func (r *Recorder) Close() {
	if r == nil {
		return
	}
	r.closeOnce.Do(func() { close(r.closed) })
	r.wg.Wait()
}

func (r *Recorder) loop() {
	defer r.wg.Done()

	flush := func() {
		for _, s := range r.sinks {
			s.flush(r.cfg.Dir)
		}
	}
	add := func(e any) {
		for _, s := range r.sinks {
			if s.add(e) {
				break
			}
		}
		for _, s := range r.sinks {
			if s.len() >= flushRows {
				flush()
				break
			}
		}
	}

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for {
		select {
		case e := <-r.ch:
			add(e)
		case <-ticker.C:
			flush()
		case <-r.closed:
			for {
				select {
				case e := <-r.ch:
					add(e)
				default:
					flush()
					return
				}
			}
		}
	}
}
