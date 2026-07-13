package analytics

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// cutoffForRequest runs cutoffFromDays through a real Fiber context built
// from a request URL, since it reads from c.QueryInt rather than taking a
// plain int — exercising the actual parsing, not just the clamp math.
func cutoffForRequest(t *testing.T, rawQuery string) time.Time {
	t.Helper()
	app := fiber.New()
	var got time.Time
	app.Get("/x", func(c *fiber.Ctx) error {
		got = cutoffFromDays(c)
		return nil
	})
	req := httptest.NewRequest("GET", "/x?"+rawQuery, nil)
	if _, err := app.Test(req, -1); err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return got
}

func daysAgo(n int) time.Time { return time.Now().UTC().AddDate(0, 0, -n) }

func withinTolerance(a, b time.Time) bool {
	d := a.Sub(b)
	if d < 0 {
		d = -d
	}
	return d < 5*time.Second // generous margin for test execution time
}

func TestCutoffFromDaysDefault(t *testing.T) {
	got := cutoffForRequest(t, "")
	if !withinTolerance(got, daysAgo(defaultDays)) {
		t.Errorf("no days param: got cutoff %v, want ~%v (30 days ago)", got, daysAgo(defaultDays))
	}
}

func TestCutoffFromDaysNormalValue(t *testing.T) {
	got := cutoffForRequest(t, "days=7")
	if !withinTolerance(got, daysAgo(7)) {
		t.Errorf("days=7: got cutoff %v, want ~%v", got, daysAgo(7))
	}
}

func TestCutoffFromDaysClampsBelowOne(t *testing.T) {
	for _, raw := range []string{"days=0", "days=-5"} {
		got := cutoffForRequest(t, raw)
		if !withinTolerance(got, daysAgo(1)) {
			t.Errorf("%s: got cutoff %v, want clamped to ~%v (1 day ago)", raw, got, daysAgo(1))
		}
	}
}

func TestCutoffFromDaysClampsAboveMax(t *testing.T) {
	got := cutoffForRequest(t, "days=99999")
	if !withinTolerance(got, daysAgo(maxDays)) {
		t.Errorf("days=99999: got cutoff %v, want clamped to ~%v (365 days ago)", got, daysAgo(maxDays))
	}
}

func TestCutoffFromDaysInvalidFallsBackToDefault(t *testing.T) {
	got := cutoffForRequest(t, "days=not-a-number")
	if !withinTolerance(got, daysAgo(defaultDays)) {
		t.Errorf("days=not-a-number: got cutoff %v, want default ~%v (30 days ago)", got, daysAgo(defaultDays))
	}
}
