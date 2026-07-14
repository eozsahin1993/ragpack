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

// SetRequestHeaders applies the identity RagPack presents to remote servers
// for both the real fetch and any pre-fetch probe (e.g. MIME-type detection).
// Kept in one place so a probe request and the fetch it's predicting for are
// never treated differently by the origin's bot detection.
func SetRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; RagPack/1.0; +https://github.com/eozsahin1993/ragpack)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
}

func (f *HTTPFetcher) Fetch(ctx context.Context, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http fetcher: build request: %w", err)
	}

	SetRequestHeaders(req)

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

// FetchConditional sends If-None-Match/If-Modified-Since (whichever the
// caller has) and reports NotModified on a 304 instead of fetching the body.
func (f *HTTPFetcher) FetchConditional(ctx context.Context, uri, etag, lastModified string) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http fetcher: build request: %w", err)
	}
	SetRequestHeaders(req)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http fetcher: request failed: %w", err)
	}
	if resp.StatusCode == http.StatusNotModified {
		resp.Body.Close()
		return &FetchResult{NotModified: true}, nil
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("http fetcher: unexpected status %d for %q", resp.StatusCode, uri)
	}
	return &FetchResult{
		Body:         resp.Body,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}, nil
}
