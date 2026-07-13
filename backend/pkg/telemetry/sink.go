package telemetry

import "ragpack/pkg/telemetry/schema"

// sink buffers one event type and flushes it to Parquet. Recorder.loop stays
// generic over event types by holding a []sink instead of one typed slice +
// switch per event kind — adding a new event type is a new schema.Table
// registration in New(), not a new case here.
type sink interface {
	add(e any) bool
	len() int
	flush(dir string)
}

type typedSink[T any] struct {
	table schema.Table[T]
	buf   []T
}

func newSink[T any](t schema.Table[T]) *typedSink[T] {
	return &typedSink[T]{table: t}
}

func (s *typedSink[T]) add(e any) bool {
	v, ok := e.(T)
	if ok {
		s.buf = append(s.buf, v)
	}
	return ok
}

func (s *typedSink[T]) len() int { return len(s.buf) }

func (s *typedSink[T]) flush(dir string) {
	if len(s.buf) == 0 {
		return
	}
	writeParquet(dir, s.table.Name, s.table.Build(s.buf))
	s.buf = s.buf[:0]
}
