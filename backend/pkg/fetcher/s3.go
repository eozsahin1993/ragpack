package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

type S3Fetcher struct {
	client *s3.Client
}

func NewS3Fetcher(ctx context.Context) (*S3Fetcher, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("s3 fetcher: load aws config: %w", err)
	}
	return &S3Fetcher{client: s3.NewFromConfig(cfg)}, nil
}

func (f *S3Fetcher) Fetch(ctx context.Context, uri string) (io.ReadCloser, error) {
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return nil, err
	}

	out, err := f.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 fetcher: get %q: %w", uri, err)
	}

	return out.Body, nil
}

// FetchConditional: unlike plain HTTP, S3 has no "200 vs 304" return value — a condition that isn't met surfaces
// as an error wrapping a 304, recovered here via smithy-go's *smithyhttp.ResponseError (every S3 SDK error is
// wrapped with one — see s3shared.AddResponseErrorMiddleware — so this is documented behavior, not a guess).
func (f *S3Fetcher) FetchConditional(ctx context.Context, uri, etag, lastModified string) (*FetchResult, error) {
	bucket, key, err := parseS3URI(uri)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}
	if etag != "" {
		input.IfNoneMatch = aws.String(etag)
	}
	if lastModified != "" {
		if t, err := http.ParseTime(lastModified); err == nil { // caller formats col.LastAutoRefreshAt the same way for both sources
			input.IfModifiedSince = aws.Time(t)
		}
	}

	out, err := f.client.GetObject(ctx, input)
	if err != nil {
		var respErr *smithyhttp.ResponseError
		if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotModified {
			return &FetchResult{NotModified: true}, nil
		}
		return nil, fmt.Errorf("s3 fetcher: get %q: %w", uri, err)
	}

	var newETag, newLastMod string
	if out.ETag != nil {
		newETag = *out.ETag
	}
	if out.LastModified != nil {
		newLastMod = out.LastModified.Format(http.TimeFormat) // store in HTTP format either way
	}
	return &FetchResult{Body: out.Body, ETag: newETag, LastModified: newLastMod}, nil
}

func parseS3URI(uri string) (bucket, key string, err error) {
	trimmed := strings.TrimPrefix(uri, "s3://")
	idx := strings.IndexByte(trimmed, '/')
	if idx == -1 {
		return "", "", fmt.Errorf("s3 fetcher: invalid URI %q (no key)", uri)
	}
	return trimmed[:idx], trimmed[idx+1:], nil
}
