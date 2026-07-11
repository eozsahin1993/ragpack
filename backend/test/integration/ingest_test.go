//go:build integration

package integration_test

import (
	"testing"

	"ragpack/test/integration/helpers"
)

func TestIngestDocument(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Ingest Test")

	helpers.UploadDoc(t, app, slug, "doc.txt", "A short document about gardening.", nil)
	doc := helpers.WaitForDocument(t, app, slug, "doc.txt")

	if doc["status"] != "complete" {
		t.Fatalf("expected status complete, got %v", doc["status"])
	}
	if doc["file_uri"] != "upload://doc.txt" {
		t.Errorf("expected file_uri upload://doc.txt, got %v", doc["file_uri"])
	}
}
