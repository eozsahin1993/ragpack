package ingest

import (
	"context"
	"net/http"
	"strings"
	"time"
)

var extMime = map[string]string{
	".pdf":      "application/pdf",
	".md":       "text/markdown",
	".markdown": "text/markdown",
	".html":     "text/html",
	".htm":      "text/html",
	".txt":      "text/plain",
}

// detectMimeType infers the MIME type for a URI.
// It checks the file extension first, then issues a HEAD request for HTTP(S) URLs.
// Falls back to "text/html" for bare HTTP(S) URLs with no recognisable extension.
func detectMimeType(ctx context.Context, uri string) string {
	lower := strings.ToLower(uri)

	for ext, mime := range extMime {
		// Match extension at end or before a query string.
		withoutQuery := strings.SplitN(lower, "?", 2)[0]
		if strings.HasSuffix(withoutQuery, ext) {
			return mime
		}
	}

	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, uri, nil)
		if err == nil {
			if resp, err := client.Do(req); err == nil {
				resp.Body.Close()
				ct := resp.Header.Get("Content-Type")
				if ct != "" {
					return strings.TrimSpace(strings.SplitN(ct, ";", 2)[0])
				}
			}
		}
		return "text/html"
	}

	return "text/plain"
}
