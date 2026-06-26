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

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; RagPack/1.0; +https://github.com/eozsahin1993/ragpack)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

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
