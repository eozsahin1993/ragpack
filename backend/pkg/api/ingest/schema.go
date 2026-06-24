package ingest

// URIRequest is used when the file already lives at a remote URI (s3://, https://, file://).
// MimeType is optional — if omitted the backend detects it from the URL extension or Content-Type header.
type URIRequest struct {
	FileURI  string `json:"file_uri" validate:"required"`
	MimeType string `json:"mime_type"`
}
