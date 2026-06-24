package chunker

import (
	"fmt"
	"strings"
)

func errBadConfig(size, overlap int) error {
	return fmt.Errorf("chunker: overlap (%d) must be less than chunk size (%d)", overlap, size)
}

// splitOversize splits a single oversized text with the sliding-window strategy
// and appends the resulting sub-chunks to dst, starting at index idx.
// Returns the new index after all sub-chunks are appended.
func splitOversize(text string, idx int, cfg Config, dst *[]Chunk) int {
	runes := []rune(text)
	total := len(runes)
	step := cfg.ChunkSize - cfg.Overlap
	if step <= 0 {
		step = cfg.ChunkSize
	}
	for pos := 0; pos < total; pos += step {
		end := pos + cfg.ChunkSize
		if end > total {
			end = total
		}
		t := strings.TrimSpace(string(runes[pos:end]))
		if t != "" {
			*dst = append(*dst, Chunk{Text: t, Index: idx})
			idx++
		}
		if end == total {
			break
		}
	}
	return idx
}

