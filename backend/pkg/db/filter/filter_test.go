package filter

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"ragpack/pkg/meta"
)

// testFieldMap is the shared fixture for all tests that need metadata fields.
var testFieldMap = FieldMap{
	"title":     {Name: "title", Type: "str", Slot: 1},
	"author":    {Name: "author", Type: "str", Slot: 2},
	"score":     {Name: "score", Type: "num", Slot: 1},
	"active":    {Name: "active", Type: "bool", Slot: 1},
	"published": {Name: "published", Type: "date", Slot: 1},
	"tags":      {Name: "tags", Type: "arr", Slot: 1},
}

func field(name, typ string, slot int) meta.MetadataField {
	return meta.MetadataField{Name: name, Type: typ, Slot: slot}
}

func compile(t *testing.T, raw string) string {
	t.Helper()
	sql, err := Compile(json.RawMessage(raw), testFieldMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return sql
}

func compileErr(t *testing.T, raw string) error {
	t.Helper()
	_, err := Compile(json.RawMessage(raw), testFieldMap)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	return err
}

// ── Empty / nil input ────────────────────────────────────────────────────────

func TestCompile_NilReturnsEmpty(t *testing.T) {
	sql, err := Compile(nil, testFieldMap)
	if err != nil || sql != "" {
		t.Fatalf("expected empty string, got %q %v", sql, err)
	}
}

func TestCompile_EmptyBytesReturnsEmpty(t *testing.T) {
	sql, err := Compile(json.RawMessage{}, testFieldMap)
	if err != nil || sql != "" {
		t.Fatalf("expected empty string, got %q %v", sql, err)
	}
}

func TestCompile_InvalidJSONErrors(t *testing.T) {
	_, err := Compile(json.RawMessage(`{bad`), testFieldMap)
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
}

// ── Built-in columns ─────────────────────────────────────────────────────────

func TestCompile_BuiltInStrField(t *testing.T) {
	sql := compile(t, `{"mime_type": "application/pdf"}`)
	if sql != "mime_type = 'application/pdf'" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_BuiltInTimestampEq(t *testing.T) {
	// created_at with an ISO date string → unix int64
	sql := compile(t, `{"created_at": {"$gte": "2024-01-01"}}`)
	if !strings.HasPrefix(sql, "created_at >= ") {
		t.Errorf("expected timestamp comparison, got: %s", sql)
	}
}

// ── Simple equality ───────────────────────────────────────────────────────────

func TestCompile_StrEq(t *testing.T) {
	sql := compile(t, `{"title": "hello"}`)
	if sql != "metadata_str_1 = 'hello'" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_NumEq(t *testing.T) {
	sql := compile(t, `{"score": 4.5}`)
	if sql != "metadata_num_1 = 4.5" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_BoolEqTrue(t *testing.T) {
	sql := compile(t, `{"active": true}`)
	if sql != "metadata_bool_1 = true" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_BoolEqFalse(t *testing.T) {
	sql := compile(t, `{"active": false}`)
	if sql != "metadata_bool_1 = false" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

// ── Comparison operators ──────────────────────────────────────────────────────

func TestCompile_NumGt(t *testing.T) {
	sql := compile(t, `{"score": {"$gt": 3}}`)
	if sql != "metadata_num_1 > 3" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_NumGte(t *testing.T) {
	sql := compile(t, `{"score": {"$gte": 3}}`)
	if sql != "metadata_num_1 >= 3" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_NumLt(t *testing.T) {
	sql := compile(t, `{"score": {"$lt": 10}}`)
	if sql != "metadata_num_1 < 10" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_NumLte(t *testing.T) {
	sql := compile(t, `{"score": {"$lte": 10}}`)
	if sql != "metadata_num_1 <= 10" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_StrNe(t *testing.T) {
	sql := compile(t, `{"title": {"$ne": "draft"}}`)
	if sql != "metadata_str_1 != 'draft'" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

// ── Range query (multiple operators on same field) ────────────────────────────

func TestCompile_NumRange(t *testing.T) {
	sql := compile(t, `{"score": {"$gte": 1, "$lte": 5}}`)
	// Parts can appear in either order (map iteration); check both are present.
	if !strings.Contains(sql, "metadata_num_1 >= 1") || !strings.Contains(sql, "metadata_num_1 <= 5") {
		t.Errorf("unexpected sql: %s", sql)
	}
	if !strings.HasPrefix(sql, "(") || !strings.HasSuffix(sql, ")") {
		t.Errorf("expected range to be parenthesised: %s", sql)
	}
}

// ── $in / $nin ────────────────────────────────────────────────────────────────

func TestCompile_StrIn(t *testing.T) {
	sql := compile(t, `{"title": {"$in": ["a", "b"]}}`)
	if sql != "metadata_str_1 IN ('a', 'b')" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_NumNin(t *testing.T) {
	sql := compile(t, `{"score": {"$nin": [1, 2, 3]}}`)
	if sql != "metadata_num_1 NOT IN (1, 2, 3)" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

// ── $exists ───────────────────────────────────────────────────────────────────

func TestCompile_ExistsTrue(t *testing.T) {
	sql := compile(t, `{"title": {"$exists": true}}`)
	if sql != "metadata_str_1 IS NOT NULL" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_ExistsFalse(t *testing.T) {
	sql := compile(t, `{"title": {"$exists": false}}`)
	if sql != "metadata_str_1 IS NULL" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

// ── $like / $ilike ────────────────────────────────────────────────────────────

func TestCompile_StrLike(t *testing.T) {
	sql := compile(t, `{"title": {"$like": "%guide%"}}`)
	if sql != "metadata_str_1 LIKE '%guide%'" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_StrIlike(t *testing.T) {
	sql := compile(t, `{"title": {"$ilike": "%Guide%"}}`)
	if sql != "metadata_str_1 ILIKE '%Guide%'" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_LikeOnNumErrors(t *testing.T) {
	compileErr(t, `{"score": {"$like": "%5%"}}`)
}

func TestCompile_IlikeOnNumErrors(t *testing.T) {
	compileErr(t, `{"score": {"$ilike": "%5%"}}`)
}

// ── Array operators ───────────────────────────────────────────────────────────

func TestCompile_ArrContains(t *testing.T) {
	sql := compile(t, `{"tags": {"$contains": "go"}}`)
	if sql != "array_has(metadata_arr_1, 'go')" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_ArrContainsAny(t *testing.T) {
	sql := compile(t, `{"tags": {"$containsAny": ["go", "rust"]}}`)
	if sql != "array_has_any(metadata_arr_1, make_array('go', 'rust'))" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_ArrContainsAll(t *testing.T) {
	sql := compile(t, `{"tags": {"$containsAll": ["go", "rust"]}}`)
	if sql != "array_has_all(metadata_arr_1, make_array('go', 'rust'))" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

func TestCompile_ContainsOnStrErrors(t *testing.T) {
	compileErr(t, `{"title": {"$contains": "x"}}`)
}

// ── Logical operators ─────────────────────────────────────────────────────────

func TestCompile_And(t *testing.T) {
	sql := compile(t, `{"$and": [{"title": "x"}, {"score": 1}]}`)
	want := "(metadata_str_1 = 'x' AND metadata_num_1 = 1)"
	if sql != want {
		t.Errorf("unexpected sql:\n got  %s\n want %s", sql, want)
	}
}

func TestCompile_Or(t *testing.T) {
	sql := compile(t, `{"$or": [{"title": "x"}, {"title": "y"}]}`)
	want := "(metadata_str_1 = 'x' OR metadata_str_1 = 'y')"
	if sql != want {
		t.Errorf("unexpected sql:\n got  %s\n want %s", sql, want)
	}
}

func TestCompile_NestedLogical(t *testing.T) {
	sql := compile(t, `{"$and": [{"$or": [{"title": "a"}, {"title": "b"}]}, {"score": {"$gt": 0}}]}`)
	want := "((metadata_str_1 = 'a' OR metadata_str_1 = 'b') AND metadata_num_1 > 0)"
	if sql != want {
		t.Errorf("unexpected sql:\n got  %s\n want %s", sql, want)
	}
}

func TestCompile_AndEmptyArrayErrors(t *testing.T) {
	compileErr(t, `{"$and": []}`)
}

// ── Multiple top-level fields (implicit AND) ──────────────────────────────────

func TestCompile_ImplicitAnd(t *testing.T) {
	sql := compile(t, `{"title": "x", "score": 1}`)
	// Two fields joined with AND
	if !strings.Contains(sql, "metadata_str_1 = 'x'") || !strings.Contains(sql, "metadata_num_1 = 1") {
		t.Errorf("unexpected sql: %s", sql)
	}
	if !strings.Contains(sql, " AND ") {
		t.Errorf("expected AND between fields: %s", sql)
	}
}

// ── Date field ────────────────────────────────────────────────────────────────

func TestCompile_DateGte(t *testing.T) {
	sql := compile(t, `{"published": {"$gte": "2024-06-01"}}`)
	// Should emit an integer unix timestamp
	if !strings.HasPrefix(sql, "metadata_date_1 >= ") {
		t.Errorf("unexpected sql: %s", sql)
	}
	rest := strings.TrimPrefix(sql, "metadata_date_1 >= ")
	if len(rest) == 0 || rest[0] == '\'' {
		t.Errorf("date should produce an int literal, got: %s", sql)
	}
}

func TestCompile_DateUnixInt(t *testing.T) {
	// Passing a unix epoch directly as a number
	sql := compile(t, `{"published": {"$gte": 1704067200}}`)
	if sql != "metadata_date_1 >= 1704067200" {
		t.Errorf("unexpected sql: %s", sql)
	}
}

// ── Unknown field / operator errors ──────────────────────────────────────────

func TestCompile_UnknownFieldErrors(t *testing.T) {
	err := compileErr(t, `{"nonexistent": "value"}`)
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCompile_UnknownOperatorErrors(t *testing.T) {
	err := compileErr(t, `{"title": {"$regex": ".*"}}`)
	if !strings.Contains(err.Error(), "unknown operator") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCompile_ExistsNonBoolErrors(t *testing.T) {
	compileErr(t, `{"title": {"$exists": 1}}`)
}

// ── SQL injection prevention ──────────────────────────────────────────────────

func TestCompile_QuoteSQLString_SingleQuote(t *testing.T) {
	result := quoteSQLString("it's a test")
	if result != "'it''s a test'" {
		t.Errorf("unexpected quoting: %s", result)
	}
}

func TestCompile_QuoteSQLString_Benign(t *testing.T) {
	result := quoteSQLString("hello world")
	if result != "'hello world'" {
		t.Errorf("unexpected quoting: %s", result)
	}
}

func TestCompile_InjectionInEq(t *testing.T) {
	// Classic injection attempt: value ends the string and appends OR 1=1
	payload := `' OR '1'='1`
	raw := fmt.Sprintf(`{"title": %q}`, payload)
	sql := compile(t, raw)
	// The payload must be safely quoted — no unescaped single quotes outside the literal
	if strings.Contains(sql, "OR '1'='1") {
		t.Errorf("SQL injection not escaped: %s", sql)
	}
	if !strings.Contains(sql, "''") {
		t.Errorf("expected doubled quote in output: %s", sql)
	}
}

func TestCompile_InjectionInLike(t *testing.T) {
	// The ' in the payload must be doubled to '' so it can't terminate the SQL literal.
	payload := `%'; DROP TABLE chunks; --`
	raw := fmt.Sprintf(`{"title": {"$like": %q}}`, payload)
	sql := compile(t, raw)
	want := `metadata_str_1 LIKE '%''; DROP TABLE chunks; --'`
	if sql != want {
		t.Errorf("unexpected sql:\n got  %s\n want %s", sql, want)
	}
}

func TestCompile_InjectionInContains(t *testing.T) {
	payload := `go'); DROP TABLE chunks; --`
	raw := fmt.Sprintf(`{"tags": {"$contains": %q}}`, payload)
	sql := compile(t, raw)
	want := `array_has(metadata_arr_1, 'go''); DROP TABLE chunks; --')`
	if sql != want {
		t.Errorf("unexpected sql:\n got  %s\n want %s", sql, want)
	}
}

func TestCompile_InjectionInIn(t *testing.T) {
	payload := `a', 'b'); DROP TABLE chunks; --`
	raw := fmt.Sprintf(`{"title": {"$in": [%q]}}`, payload)
	sql := compile(t, raw)
	want := `metadata_str_1 IN ('a'', ''b''); DROP TABLE chunks; --')`
	if sql != want {
		t.Errorf("unexpected sql:\n got  %s\n want %s", sql, want)
	}
}

// ── resolveColumn ─────────────────────────────────────────────────────────────

func TestResolveColumn_BuiltInTimestamp(t *testing.T) {
	col, typ, err := resolveColumn("created_at", nil)
	if err != nil || col != "created_at" || typ != "timestamp" {
		t.Errorf("got col=%s typ=%s err=%v", col, typ, err)
	}
}

func TestResolveColumn_BuiltInStr(t *testing.T) {
	for _, name := range []string{"mime_type", "source_name", "external_id", "file_uri"} {
		col, typ, err := resolveColumn(name, nil)
		if err != nil || col != name || typ != "str" {
			t.Errorf("field %s: got col=%s typ=%s err=%v", name, col, typ, err)
		}
	}
}

func TestResolveColumn_MetadataField(t *testing.T) {
	col, typ, err := resolveColumn("score", testFieldMap)
	if err != nil || col != "metadata_num_1" || typ != "num" {
		t.Errorf("got col=%s typ=%s err=%v", col, typ, err)
	}
}

func TestResolveColumn_UnknownErrors(t *testing.T) {
	_, _, err := resolveColumn("ghost", testFieldMap)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

// ── slotColumn ────────────────────────────────────────────────────────────────

func TestSlotColumn(t *testing.T) {
	cases := []struct {
		typ  string
		slot int
		want string
	}{
		{"str", 1, "metadata_str_1"},
		{"num", 3, "metadata_num_3"},
		{"bool", 2, "metadata_bool_2"},
		{"date", 5, "metadata_date_5"},
		{"arr", 10, "metadata_arr_10"},
	}
	for _, tc := range cases {
		got := slotColumn(tc.typ, tc.slot)
		if got != tc.want {
			t.Errorf("slotColumn(%q, %d) = %q, want %q", tc.typ, tc.slot, got, tc.want)
		}
	}
}

// ── formatTimestamp relative dates ────────────────────────────────────────────

func TestParseRelativeDate_Today(t *testing.T) {
	ts, err := parseRelativeDate("today")
	if err != nil || ts.IsZero() {
		t.Errorf("unexpected: %v %v", ts, err)
	}
}

func TestParseRelativeDate_NDaysAgo(t *testing.T) {
	ts, err := parseRelativeDate("7 days ago")
	if err != nil || ts.IsZero() {
		t.Errorf("unexpected: %v %v", ts, err)
	}
}

func TestParseRelativeDate_Unknown(t *testing.T) {
	_, err := parseRelativeDate("next tuesday")
	if err == nil {
		t.Fatal("expected error for unrecognized relative date")
	}
}
