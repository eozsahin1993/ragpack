package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http fetcher: build request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http fetcher: request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("http fetcher: unexpected status %d for %q", resp.StatusCode, uri)
	}

	return resp.Body, nil
}
