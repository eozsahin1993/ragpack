package util

// NonEmptyStr returns a pointer to s if s is non-empty, otherwise nil.
func NonEmptyStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
