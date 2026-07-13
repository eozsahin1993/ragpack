package schema

import (
	"testing"
	"time"
)

func TestStampFillsEmptyIDAndZeroTime(t *testing.T) {
	var id string
	var at time.Time
	Stamp(&id, &at)

	if id == "" {
		t.Error("want a generated id, got empty string")
	}
	if at.IsZero() {
		t.Error("want a generated timestamp, got zero value")
	}
}

func TestStampPreservesExistingValues(t *testing.T) {
	id := "existing-id"
	at := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	Stamp(&id, &at)

	if id != "existing-id" {
		t.Errorf("id changed: got %q, want %q", id, "existing-id")
	}
	if !at.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("timestamp changed: got %v", at)
	}
}
