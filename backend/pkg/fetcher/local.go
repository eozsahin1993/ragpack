package fetcher

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

type LocalFetcher struct{}

func NewLocalFetcher() *LocalFetcher {
	return &LocalFetcher{}
}

func (f *LocalFetcher) Fetch(_ context.Context, uri string) (io.ReadCloser, error) {
	path := strings.TrimPrefix(uri, "file://")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("local fetcher: open %q: %w", path, err)
	}
	return file, nil
}
