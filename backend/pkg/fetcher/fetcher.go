package fetcher

import (
	"context"
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
	default:
		return NewLocalFetcher(), nil
	}
}
