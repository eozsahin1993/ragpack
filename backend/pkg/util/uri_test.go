package util

import "testing"

func TestNameFromURI(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"upload://my-file.pdf", "my-file"},
		{"upload://report.docx", "report"},
		{"upload://notes.txt", "notes"},
		{"https://example.com/article.html", "article"},
		{"https://example.com/path/to/doc.pdf", "doc"},
		{"upload://no-extension", "no-extension"},
		{"upload://", ""},
		{"", ""},
		{"upload://file.tar.gz", "file.tar"},
	}

	for _, tt := range tests {
		got := NameFromURI(tt.uri)
		if got != tt.want {
			t.Errorf("NameFromURI(%q) = %q, want %q", tt.uri, got, tt.want)
		}
	}
}
