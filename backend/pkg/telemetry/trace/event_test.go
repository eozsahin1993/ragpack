package trace

import (
	"testing"
	"time"
)

func TestPrepareDropsWhenEventIDEmpty(t *testing.T) {
	e := Event{}
	if ok := e.Prepare(false); ok {
		t.Fatal("want drop when EventID (the FK to query_events) is unset")
	}
}

func TestPrepareDropsWhenRedacted(t *testing.T) {
	e := Event{EventID: "q1"}
	if ok := e.Prepare(true); ok {
		t.Fatal("want drop under redaction, even with a valid EventID")
	}
}

func TestPrepareNeverGeneratesEventID(t *testing.T) {
	e := Event{EventID: "q1"}
	e.Prepare(false)
	if e.EventID != "q1" {
		t.Errorf("EventID changed: got %q, want unchanged", e.EventID)
	}
}

func TestPrepareStampsOccurredAtWhenUnset(t *testing.T) {
	e := Event{EventID: "q1"}
	if ok := e.Prepare(false); !ok {
		t.Fatal("want Prepare to keep the row")
	}
	if e.OccurredAt.IsZero() {
		t.Error("want a stamped OccurredAt")
	}
}

func TestPreparePreservesExistingOccurredAt(t *testing.T) {
	at := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	e := Event{EventID: "q1", OccurredAt: at}

	e.Prepare(false)

	if !e.OccurredAt.Equal(at) {
		t.Errorf("OccurredAt changed: got %v", e.OccurredAt)
	}
}
