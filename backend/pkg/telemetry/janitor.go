package telemetry

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const janitorInterval = time.Hour

// janitor enforces retention and the size cap. Parquet files are immutable,
// so enforcement is per-file: whole rotated files are deleted, never rows.
func (r *Recorder) janitor() {
	defer r.wg.Done()
	r.enforce()
	ticker := time.NewTicker(janitorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.closed:
			return
		case <-ticker.C:
			r.enforce()
		}
	}
}

func (r *Recorder) enforce() {
	type parquetFile struct {
		path string
		mod  time.Time
		size int64
	}

	var files []parquetFile
	filepath.WalkDir(r.cfg.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".parquet") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		files = append(files, parquetFile{path: path, mod: info.ModTime(), size: info.Size()})
		return nil
	})

	cutoff := time.Now().Add(-time.Duration(r.cfg.RetentionDays) * 24 * time.Hour)
	var total int64
	removed := 0
	kept := files[:0]
	for _, f := range files {
		if f.mod.Before(cutoff) {
			if err := os.Remove(f.path); err == nil {
				removed++
			}
			continue
		}
		kept = append(kept, f)
		total += f.size
	}

	sort.Slice(kept, func(i, j int) bool { return kept[i].mod.Before(kept[j].mod) })
	cap := int64(r.cfg.MaxSizeMB) * 1024 * 1024
	for i := 0; total > cap && i < len(kept); i++ {
		if err := os.Remove(kept[i].path); err == nil {
			total -= kept[i].size
			removed++
		}
	}

	if removed > 0 {
		log.Printf("telemetry: janitor removed %d parquet files (retention %dd, cap %dMB)",
			removed, r.cfg.RetentionDays, r.cfg.MaxSizeMB)
	}
}
