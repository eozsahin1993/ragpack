package telemetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"

	"ragpack/pkg/telemetry/schema"
)

type fakeEvent struct {
	ID string
}

var fakeSchema = arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.BinaryTypes.String}}, nil)

var fakeTable = schema.Table[*fakeEvent]{
	Name:   "fake_events",
	Schema: fakeSchema,
	Build: func(events []*fakeEvent) arrow.Record {
		b := array.NewRecordBuilder(memory.DefaultAllocator, fakeSchema)
		defer b.Release()
		for _, e := range events {
			b.Field(0).(*array.StringBuilder).Append(e.ID)
		}
		return b.NewRecord()
	},
}

func TestTypedSinkAddRejectsMismatchedType(t *testing.T) {
	s := newSink(fakeTable)
	if s.add("not a *fakeEvent") {
		t.Fatal("want add to reject a value of the wrong type")
	}
	if s.len() != 0 {
		t.Fatalf("want len 0 after a rejected add, got %d", s.len())
	}
}

func TestTypedSinkAddAcceptsMatchingType(t *testing.T) {
	s := newSink(fakeTable)
	if !s.add(&fakeEvent{ID: "a"}) {
		t.Fatal("want add to accept a *fakeEvent")
	}
	if s.len() != 1 {
		t.Fatalf("want len 1, got %d", s.len())
	}
}

func TestTypedSinkFlushIsNoopWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	s := newSink(fakeTable)
	s.flush(dir)
	if _, err := os.Stat(filepath.Join(dir, "fake_events")); !os.IsNotExist(err) {
		t.Error("want no output directory created for an empty flush")
	}
}

func TestTypedSinkFlushWritesFileAndResetsBuffer(t *testing.T) {
	dir := t.TempDir()
	s := newSink(fakeTable)
	s.add(&fakeEvent{ID: "a"})
	s.add(&fakeEvent{ID: "b"})

	s.flush(dir)

	if s.len() != 0 {
		t.Fatalf("want buffer reset after flush, got len %d", s.len())
	}
	files := parquetFiles(t, filepath.Join(dir, "fake_events"))
	if len(files) != 1 {
		t.Fatalf("want 1 parquet file, got %d", len(files))
	}
}
