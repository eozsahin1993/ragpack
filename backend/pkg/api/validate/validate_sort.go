package validate

import (
	"fmt"
	"slices"

	"github.com/go-playground/validator/v10"

	"ragpack/pkg/meta"
)

var sortValidatorSpecs = map[string]meta.SortSpec{}

// RegisterSortValidator wires a `validate:"<tag>"` rule to spec, so any
// entity's SortSpec (see meta.Sortable) can back a query-param validator
// with one call instead of a hand-written closure per entity. It also
// records spec under tag, so errorMessage can report the offending value
// and the valid options without a case per registered tag.
func RegisterSortValidator(tag string, spec meta.SortSpec) {
	v.RegisterValidation(tag, func(fl validator.FieldLevel) bool { //nolint:errcheck
		return spec.IsValid(fl.Field().String())
	})
	sortValidatorSpecs[tag] = spec
}

func sortValidatorMessage(fe validator.FieldError) (string, bool) {
	spec, ok := sortValidatorSpecs[fe.Tag()]
	if !ok {
		return "", false
	}
	options := make([]string, 0, len(spec.Valid))
	for field := range spec.Valid {
		options = append(options, field)
	}
	slices.Sort(options)
	return fmt.Sprintf("%q is not a valid sort field, options are %v", fe.Value(), options), true
}
