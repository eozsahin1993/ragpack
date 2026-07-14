package chunker

import (
	"crypto/sha256"
	"fmt"
)

// Version: bump alongside any parser/chunker code change that could alter output for identical input — never reuse across pipeline versions.
const Version = "1"

// Fingerprint hashes pipeline version + chunk config + the literal text sent to the embedder (caller passes header-prefixed text when a chunk has one).
func Fingerprint(cfg Config, embeddedText string) string {
	input := fmt.Sprintf("%s|%s|%d|%d|%s", Version, cfg.Strategy, cfg.ChunkSize, cfg.Overlap, embeddedText)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}
