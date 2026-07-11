package helpers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UploadDoc posts a real multipart file upload to /ingest. metadata is
// optional (pass nil to omit the field entirely).
func UploadDoc(t *testing.T, app *fiber.App, slug, filename, content string, metadata map[string]string) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	part.Write([]byte(content))
	if metadata != nil {
		metaJSON, _ := json.Marshal(metadata)
		w.WriteField("metadata", string(metaJSON))
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/admin/collections/"+slug+"/ingest", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("upload %s: %v", filename, err)
	}
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("upload %s: expected 202, got %d", filename, resp.StatusCode)
	}
	var job map[string]any
	json.NewDecoder(resp.Body).Decode(&job)
	return job
}

// WaitForDocument polls the collection's document list for the row matching
// filename's upload:// file_uri until it leaves "ingesting", or times out.
// Jobs don't carry a document_id back-reference (Document.job_id points the
// other way — see CLAUDE.md's ingestion pipeline notes), so matching by the
// upload's file_uri is the way to find the resulting document.
func WaitForDocument(t *testing.T, app *fiber.App, slug, filename string) map[string]any {
	t.Helper()
	wantURI := "upload://" + filename
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, list := DoJSON(t, app, http.MethodGet, "/admin/collections/"+slug+"/documents", nil)
		if resp.StatusCode == fiber.StatusOK {
			if docs, ok := list["documents"].([]any); ok {
				for _, d := range docs {
					doc, ok := d.(map[string]any)
					if !ok || doc["file_uri"] != wantURI {
						continue
					}
					if status, _ := doc["status"].(string); status == "complete" || status == "failed" {
						return doc
					}
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q to produce a completed document", wantURI)
	return nil
}
