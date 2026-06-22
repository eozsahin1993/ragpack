package fetcher

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func parseS3URI(uri string) (bucket, key string, err error) {
	trimmed := strings.TrimPrefix(uri, "s3://")
	idx := strings.IndexByte(trimmed, '/')
	if idx == -1 {
		return "", "", fmt.Errorf("s3 fetcher: invalid URI %q (no key)", uri)
	}
	return trimmed[:idx], trimmed[idx+1:], nil
}
