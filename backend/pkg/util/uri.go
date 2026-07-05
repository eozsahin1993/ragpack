package util

import (
	"path"
	"strings"
)

// NameFromURI derives a human-readable name from a URI by taking the last path
// segment and stripping the file extension. Returns "" if nothing useful remains.
func NameFromURI(uri string) string {
	base := path.Base(strings.TrimPrefix(uri, "upload://"))
	if ext := path.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	if base == "." {
		return ""
	}
	return base
}
