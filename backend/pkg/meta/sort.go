package meta

import "reflect"

// SortDir is a generic ascending/descending sort direction, shared across
// any store's list/sort options (not specific to one entity).
type SortDir string

const (
	SortAsc  SortDir = "asc"
	SortDesc SortDir = "desc"
)

// SortSpec is the set of columns type T allows sorting by, derived from
// which struct fields carry a `sort:"true"` or `sort:"default"` tag (column
// name taken from that field's `db` tag). One exactly-default field per
// struct is expected; the last one wins if more than one is tagged.
type SortSpec struct {
	Valid   map[string]bool
	Default string
}

func (s SortSpec) IsValid(field string) bool {
	return s.Valid[field]
}

// Sortable builds a SortSpec for T by scanning its struct tags, so any store
// type can opt into sortability the same way — via `sort:"true"`/`sort:"default"`
// tags on its fields — instead of hand-writing a whitelist per type.
func Sortable[T any]() SortSpec {
	spec := SortSpec{Valid: map[string]bool{}}
	for field := range reflect.TypeFor[T]().Fields() {
		tag := field.Tag.Get("sort")
		if tag == "" {
			continue
		}
		col := field.Tag.Get("db")
		spec.Valid[col] = true
		if tag == "default" {
			spec.Default = col
		}
	}
	return spec
}
