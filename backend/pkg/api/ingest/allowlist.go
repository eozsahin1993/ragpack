package ingest

import (
	"fmt"
	"path/filepath"
	"strings"
)

var allowedMIMETypes = map[string]struct{}{
	"text/plain":       {},
	"text/markdown":    {},
	"text/html":        {},
	"text/csv":         {},
	"text/xml":         {},
	"application/json": {},
	"application/xml":  {},
	"application/pdf":                                                             {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
}

var allowedExtensions = map[string]struct{}{
	".txt":      {},
	".md":       {},
	".markdown": {},
	".html":     {},
	".htm":      {},
	".pdf":      {},
	".docx":     {},
	".pptx":     {},
	".xlsx":     {},
	".csv":      {},
	".json":     {},
	".xml":      {},
}

// validateFile checks both the file extension and MIME type against the allowlists.
// If filename has no extension (e.g. a bare URL), only the MIME type is checked.
func validateFile(filename, mimeType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		if _, ok := allowedExtensions[ext]; !ok {
			return fmt.Errorf("unsupported file extension: %s", ext)
		}
	}
	base := strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
	if _, ok := allowedMIMETypes[base]; !ok {
		return fmt.Errorf("unsupported file type: %s", base)
	}
	return nil
}

// validateURI whitelists allowed URI schemes. file:// is intentionally excluded
// to prevent reading arbitrary files from the server filesystem.
func validateURI(uri string) bool {
	lower := strings.ToLower(uri)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "s3://")
}
