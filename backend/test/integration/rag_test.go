//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ragpack/test/integration/helpers"
)

func TestRagProducesAnswerAndSourceChunks(t *testing.T) {
	app := helpers.NewTestApp(t)
	slug := helpers.CreateCollection(t, app, "Rag Test")

	helpers.UploadDoc(t, app, slug, "platner.txt", "Graham Platner is a candidate whose campaign gained attention.", nil)
	helpers.WaitForDocument(t, app, slug, "platner.txt")

	resp, rag := helpers.DoJSON(t, app, http.MethodPost, "/admin/collections/"+slug+"/rag", map[string]any{
		"query": "Platner campaign", "top_k": 2,
	})
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("rag: got %d: %v", resp.StatusCode, rag)
	}
	if answer, _ := rag["answer"].(string); answer == "" {
		t.Errorf("expected non-empty answer")
	}
	chunks, _ := rag["chunks"].([]any)
	if len(chunks) == 0 {
		t.Errorf("expected source chunks")
	}
}
