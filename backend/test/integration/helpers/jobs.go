package helpers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// IngestURI posts a URI-based ingest request (as opposed to UploadDoc's
// multipart upload) — the path force-refresh and auto-refresh exercise,
// since upload:// documents can't be refreshed. Returns the created Job.
func IngestURI(t *testing.T, app *fiber.App, slug, fileURI, mimeType string, refresh, force bool) map[string]any {
	t.Helper()
	path := "/admin/collections/" + slug + "/ingest"
	if refresh {
		path += "?refresh=true"
		if force {
			path += "&force=true"
		}
	} else if force {
		path += "?force=true"
	}
	body := map[string]any{"file_uri": fileURI}
	if mimeType != "" {
		body["mime_type"] = mimeType
	}
	resp, job := DoJSON(t, app, http.MethodPost, path, body)
	if resp.StatusCode != fiber.StatusAccepted && resp.StatusCode != fiber.StatusOK {
		t.Fatalf("ingest %s: expected 200/202, got %d: %v", fileURI, resp.StatusCode, job)
	}
	return job
}

// WaitForJob polls a job by ID until it leaves pending/processing, or times
// out. Waiting on the job (not the document) avoids the race where a
// refresh's document briefly still shows the *previous* job's "complete"
// status before the new job has even been dequeued.
func WaitForJob(t *testing.T, app *fiber.App, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, job := DoJSON(t, app, http.MethodGet, "/admin/jobs/"+jobID, nil)
		if resp.StatusCode == fiber.StatusOK {
			if status, _ := job["status"].(string); status == "complete" || status == "failed" {
				return job
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for job %q to finish", jobID)
	return nil
}

// WaitForDocumentByURI is WaitForDocument generalized to any file_uri scheme,
// not just upload://.
func WaitForDocumentByURI(t *testing.T, app *fiber.App, slug, fileURI string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, list := DoJSON(t, app, http.MethodGet, "/admin/collections/"+slug+"/documents", nil)
		if resp.StatusCode == fiber.StatusOK {
			if docs, ok := list["documents"].([]any); ok {
				for _, d := range docs {
					doc, ok := d.(map[string]any)
					if !ok || doc["file_uri"] != fileURI {
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
	t.Fatalf("timed out waiting for %q to produce a completed document", fileURI)
	return nil
}
