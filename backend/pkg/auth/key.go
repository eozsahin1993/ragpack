package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const prefix = "rp_"

// Generate creates a new API key. Returns the plaintext (shown once), its
// SHA-256 hash (stored in SQLite), and a 4-char hint for display.
func Generate() (plaintext, hash, hint string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("auth: generate key: %w", err)
	}
	plaintext = prefix + hex.EncodeToString(b)
	hash = Hash(plaintext)
	hint = plaintext[len(plaintext)-4:]
	return
}

// Hash returns the SHA-256 hex digest of a plaintext key.
func Hash(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// Hint returns the last 4 characters of a plaintext key for display.
func Hint(plaintext string) string {
	if len(plaintext) < 4 {
		return plaintext
	}
	return plaintext[len(plaintext)-4:]
}
