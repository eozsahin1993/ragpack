package fetcher

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Fetcher interface {
	Fetch(ctx context.Context, uri string) (io.ReadCloser, error)
}

// ConditionalFetcher is an optional capability (http(s):// and s3:// implement it); the refresh scheduler type-asserts for it.
type ConditionalFetcher interface {
	FetchConditional(ctx context.Context, uri, etag, lastModified string) (*FetchResult, error)
}

// FetchResult is the outcome of a conditional fetch. Body is nil when
// NotModified is true — nothing to close, nothing to read.
type FetchResult struct {
	Body         io.ReadCloser
	NotModified  bool
	ETag         string
	LastModified string
}

// New returns the appropriate Fetcher for the given URI scheme.
// S3 fetcher loads credentials from the default AWS credential chain
// (env vars, ~/.aws/credentials, IAM role).
func New(ctx context.Context, uri string) (Fetcher, error) {
	switch {
	case strings.HasPrefix(uri, "s3://"):
		return NewS3Fetcher(ctx)
	case strings.HasPrefix(uri, "http://"), strings.HasPrefix(uri, "https://"):
		return NewHTTPFetcher(), nil
	case strings.HasPrefix(uri, "upload://"):
		return nil, fmt.Errorf("uploaded files are streamed at ingest time and cannot be re-fetched; re-submit the file")
	default:
		return NewLocalFetcher(), nil
	}
}
