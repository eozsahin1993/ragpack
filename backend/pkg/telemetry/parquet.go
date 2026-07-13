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
func writeParquet(dir, table string, rec arrow.Record) {
	defer rec.Release()

	day := time.Now().UTC().Format("2006-01-02")
	outDir := filepath.Join(dir, table, day)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Printf("telemetry: mkdir %s: %v", outDir, err)
		return
	}
	path := filepath.Join(outDir, fmt.Sprintf("%d.parquet", time.Now().UnixNano()))

	f, err := os.Create(path)
	if err != nil {
		log.Printf("telemetry: create %s: %v", path, err)
		return
	}
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Zstd))
	w, err := pqarrow.NewFileWriter(rec.Schema(), f, props, pqarrow.DefaultWriterProps())
	if err != nil {
		log.Printf("telemetry: parquet writer %s: %v", path, err)
		f.Close()
		return
	}
	if err := w.Write(rec); err != nil {
		log.Printf("telemetry: write %s: %v", path, err)
	}
	if err := w.Close(); err != nil {
		log.Printf("telemetry: close %s: %v", path, err)
	}
	f.Close()
}
