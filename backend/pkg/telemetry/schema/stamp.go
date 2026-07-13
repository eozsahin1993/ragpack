package schema

import (
	"time"

	"github.com/google/uuid"
)

// Stamp fills id/at with generated defaults if empty/zero — shared by every
// event type's Prepare method that wants a server-generated identity.
func Stamp(id *string, at *time.Time) {
	if *id == "" {
		*id = uuid.NewString()
	}
	if at.IsZero() {
		*at = time.Now().UTC()
	}
}
