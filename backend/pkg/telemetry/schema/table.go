// Package schema holds the pieces every telemetry event package needs to
// describe its own Parquet shape, without those packages depending on the
// recorder (which depends on them) — the cycle-breaking leaf.
package schema

import (
	"time"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
)

// Table describes one event type's on-disk shape: its Parquet table/folder
// name, Arrow schema, and how to turn a batch of typed events into a record.
type Table[T any] struct {
	Name   string
	Schema *arrow.Schema
	Build  func([]T) arrow.Record
}

var TsType = arrow.FixedWidthTypes.Timestamp_ms

func Ts(t time.Time) arrow.Timestamp { return arrow.Timestamp(t.UTC().UnixMilli()) }

func OptStr(b *array.StringBuilder, v *string) {
	if v == nil {
		b.AppendNull()
	} else {
		b.Append(*v)
	}
}

func OptI64(b *array.Int64Builder, v *int64) {
	if v == nil {
		b.AppendNull()
	} else {
		b.Append(*v)
	}
}

func OptF64(b *array.Float64Builder, v *float64) {
	if v == nil {
		b.AppendNull()
	} else {
		b.Append(*v)
	}
}
