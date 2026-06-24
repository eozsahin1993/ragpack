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
