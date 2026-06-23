package ingest

// URIRequest is used when the file already lives at a remote URI (s3://, https://, file://).
type URIRequest struct {
	FileURI  string `json:"file_uri"  validate:"required"`
	MimeType string `json:"mime_type" validate:"required"`
}
