package analytics

import "database/sql"

// ConnForTest exposes the underlying DuckDB connection to engine_test.go
// (package analytics_test) for introspection assertions (PRAGMA values, raw
// timeout behavior) that don't fit through the five named query methods.
// Only compiled for tests — not part of the real public API. Named to avoid
// Go's TestXxx(t *testing.T) discovery pattern, which "TestDB" collided with.
func ConnForTest(e *Engine) *sql.DB { return e.db }
