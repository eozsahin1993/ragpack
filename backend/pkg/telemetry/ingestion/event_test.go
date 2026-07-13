package ingestion

import (
	"testing"
	"time"
)

func TestPrepareStampsEmptyFields(t *testing.T) {
	e := Event{}
	if ok := e.Prepare(false); !ok {
		t.Fatal("want Prepare to keep the row")
	}
	if e.EventID == "" {
		t.Error("want a stamped EventID")
	}
	if e.OccurredAt.IsZero() {
		t.Error("want a stamped OccurredAt")
	}
}

func TestPreparePreservesExistingFields(t *testing.T) {
	at := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	e := Event{EventID: "fixed-id", OccurredAt: at}

	e.Prepare(false)

	if e.EventID != "fixed-id" {
		t.Errorf("EventID changed: got %q", e.EventID)
	}
	if !e.OccurredAt.Equal(at) {
		t.Errorf("OccurredAt changed: got %v", e.OccurredAt)
	}
}

func TestPrepareIgnoresRedact(t *testing.T) {
	e := Event{}
	if ok := e.Prepare(true); !ok {
		t.Fatal("ingestion events must never be dropped by redaction")
	}
}
