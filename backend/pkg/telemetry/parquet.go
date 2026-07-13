package telemetry

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/parquet"
	"github.com/apache/arrow/go/v17/parquet/compress"
	"github.com/apache/arrow/go/v17/parquet/pqarrow"
)

// writeParquet writes one immutable file per flush:
// <dir>/<table>/<YYYY-MM-DD>/<unixnano>.parquet
//
// The file is built under a .tmp name and moved into place with os.Rename
// only once fully written. analytics.Store's read_parquet glob re-expands on
// every dashboard query (see queries/views.go), so a query can race a flush;
// writing in place let it pick up a file with no footer yet ("no magic byte
// found") — rename is atomic, so the glob only ever sees a complete file.
func writeParquet(dir, table string, rec arrow.Record) {
	defer rec.Release()

	day := time.Now().UTC().Format("2006-01-02")
	outDir := filepath.Join(dir, table, day)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Printf("telemetry: mkdir %s: %v", outDir, err)
		return
	}
	finalPath := filepath.Join(outDir, fmt.Sprintf("%d.parquet", time.Now().UnixNano()))
	tmpPath := finalPath + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		log.Printf("telemetry: create %s: %v", tmpPath, err)
		return
	}
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Zstd))
	w, err := pqarrow.NewFileWriter(rec.Schema(), f, props, pqarrow.DefaultWriterProps())
	if err != nil {
		log.Printf("telemetry: parquet writer %s: %v", tmpPath, err)
		f.Close()
		os.Remove(tmpPath)
		return
	}
	if err := w.Write(rec); err != nil {
		log.Printf("telemetry: write %s: %v", tmpPath, err)
		w.Close()
		f.Close()
		os.Remove(tmpPath)
		return
	}
	if err := w.Close(); err != nil {
		log.Printf("telemetry: close %s: %v", tmpPath, err)
		f.Close()
		os.Remove(tmpPath)
		return
	}
	f.Close()

	if err := os.Rename(tmpPath, finalPath); err != nil {
		log.Printf("telemetry: rename %s -> %s: %v", tmpPath, finalPath, err)
		os.Remove(tmpPath)
	}
}
