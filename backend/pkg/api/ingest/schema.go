package ingest

import (
	"encoding/json"
	"fmt"
)

// URIRequest is used when the file already lives at a remote URI (s3://, https://, file://).
// MimeType is optional — if omitted the backend detects it from the URL extension or Content-Type header.
// ExtraJSON is an optional JSON blob that is stored on the document and every chunk, returned in search results.
// Metadata is an optional map of key-value pairs for metadata filtering. Keys must be pre-registered as metadata fields.
type URIRequest struct {
	FileURI   string                 `json:"file_uri"   validate:"required"`
	MimeType  string                 `json:"mime_type"`
	ExtraJSON *string                `json:"extra_json"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func (r *URIRequest) Validate() error {
	if r.FileURI == "" {
		return fmt.Errorf("file_uri is required")
	}
	if !validateURI(r.FileURI) {
		return fmt.Errorf("unsupported URI scheme — use https://, http://, or s3://")
	}
	if r.ExtraJSON != nil && !json.Valid([]byte(*r.ExtraJSON)) {
		return fmt.Errorf("extra_json must be valid JSON")
	}
	return nil
}

func isValidJSON(s *string) bool {
	if s == nil {
		return true
	}
	return json.Valid([]byte(*s))
}
