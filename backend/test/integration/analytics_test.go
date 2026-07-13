//go:build integration

// Exercises the admin-only analytics routes end to end (real HTTP, real
// wiring through pkg/app.New) — not query correctness, which is already
// covered thoroughly by pkg/analytics/engine_test.go against real Parquet
// files. What's worth proving here is the HTTP layer itself: the routes are
// actually reachable, they return the right JSON envelope shape, and a
// fresh install with zero telemetry data returns empty arrays rather than
// an error — telemetry writes are async (flushed every 60s or 500 rows, or
// on Close), so a test can't easily force real ingest/query activity to
// show up without either a long wait or an early Recorder.Close(), and the
// empty-database path is the one that would most obviously break silently.
package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestAnalyticsEndpointsReturnEmptyOnFreshInstall(t *testing.T) {
	app := helpers.NewTestApp(t)

	cases := []struct {
		path string
		key  string
	}{
		{"/admin/analytics/volume", "points"},
		{"/admin/analytics/cost-by-collection", "collections"},
		{"/admin/analytics/latency", "buckets"},
		{"/admin/analytics/ingestion-success-rate", "mime_types"},
		{"/admin/analytics/token-usage", "collections"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			resp, body := helpers.DoJSON(t, app, http.MethodGet, tc.path, nil)
			if resp.StatusCode != fiber.StatusOK {
				t.Fatalf("%s: got status %d, body %v", tc.path, resp.StatusCode, body)
			}
			items, ok := body[tc.key].([]any)
			if !ok {
				t.Fatalf("%s: want JSON array under key %q, got %v", tc.path, tc.key, body)
			}
			if len(items) != 0 {
				t.Errorf("%s: want empty array on a fresh install, got %v", tc.path, items)
			}
		})
	}
}

func TestAnalyticsEndpointsAcceptDaysParam(t *testing.T) {
	app := helpers.NewTestApp(t)

	for _, days := range []string{"1", "365", "0", "99999", "not-a-number"} {
		resp, body := helpers.DoJSON(t, app, http.MethodGet, "/admin/analytics/volume?days="+days, nil)
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("days=%s: got status %d, body %v", days, resp.StatusCode, body)
		}
	}
}
