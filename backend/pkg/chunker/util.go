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
// Chunk boundaries are snapped to word boundaries so no chunk starts or ends mid-word.
// Returns the new index after all sub-chunks are appended.
func splitOversize(text string, idx int, cfg Config, header *string, dst *[]Chunk) int {
	runes := []rune(text)
	total := len(runes)
	step := cfg.ChunkSize - cfg.Overlap
	if step <= 0 {
		step = cfg.ChunkSize
	}

	start := 0
	for start < total {
		end := start + cfg.ChunkSize
		if end >= total {
			end = total
		} else {
			end = snapWordEnd(runes, end)
		}

		t := strings.TrimSpace(string(runes[start:end]))
		if t != "" {
			*dst = append(*dst, Chunk{Text: t, Index: idx, Header: header})
			idx++
		}

		if end >= total {
			break
		}

		next := snapWordStart(runes, end-cfg.Overlap)
		if next <= start {
			next = start + 1 // guard against infinite loop on a single long token
		}
		start = next
	}
	return idx
}

// snapWordEnd walks backward from pos to the nearest word boundary so chunks
// don't end mid-word. Returns pos unchanged if no boundary is found (single long token).
func snapWordEnd(runes []rune, pos int) int {
	i := pos
	for i > 0 && !isWordBoundary(runes[i-1]) {
		i--
	}
	if i == 0 {
		return pos
	}
	return i
}

// snapWordStart walks forward from pos to the first non-whitespace rune so
// chunks don't start with leading spaces or mid-word overlap artefacts.
func snapWordStart(runes []rune, pos int) int {
	if pos < 0 {
		pos = 0
	}
	for pos < len(runes) && isWordBoundary(runes[pos]) {
		pos++
	}
	return pos
}

func isWordBoundary(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t'
}
