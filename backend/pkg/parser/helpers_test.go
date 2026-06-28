package parser

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func collect(t *testing.T, parser Parser, input string) []Unit {
	t.Helper()
	var units []Unit
	reader := io.NopCloser(strings.NewReader(input))
	for unit, err := range parser.Parse(context.Background(), reader) {
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		units = append(units, unit)
	}
	return units
}

func collectBytes(t *testing.T, parser Parser, data []byte) []Unit {
	t.Helper()
	var units []Unit
	reader := io.NopCloser(bytes.NewReader(data))
	for unit, err := range parser.Parse(context.Background(), reader) {
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		units = append(units, unit)
	}
	return units
}
