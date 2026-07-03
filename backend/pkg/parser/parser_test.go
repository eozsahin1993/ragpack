package parser

import "testing"

func TestNew_KnownTypes(t *testing.T) {
	types := []string{
		"text/plain",
		"text/markdown",
		"text/html",
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/csv",
		"application/json",
		"application/xml",
		"text/xml",
	}
	for _, mimeType := range types {
		result, err := New(mimeType)
		if err != nil {
			t.Errorf("New(%q) returned error: %v", mimeType, err)
		}
		if result == nil {
			t.Errorf("New(%q) returned nil parser", mimeType)
		}
	}
}

func TestNew_UnknownType(t *testing.T) {
	_, err := New("application/octet-stream")
	if err == nil {
		t.Error("want error for unknown mime type, got nil")
	}
}
