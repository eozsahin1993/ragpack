package query

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta     meta.MetaStore
	vectorDb db.VectorDb
	registry *embedder.Registry
}

func NewHandler(ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry) *Handler {
	return &Handler{meta: ms, vectorDb: vec, registry: registry}
}

func (h *Handler) Query(c *fiber.Ctx) error {
	collection, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	var req QueryRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}
	if req.TopK == 0 {
		req.TopK = 10
	}

	emb, err := h.registry.Get(collection.EmbedModel)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no embedder available for this collection's model"})
	}

	vectors, err := emb.Embed(c.Context(), []string{req.Query})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to embed query"})
	}

	results, err := h.vectorDb.QuerySimilarVectors(c.Context(), collection.TableName, embedder.Normalize(vectors[0]), req.TopK)
	if err != nil {
		log.Printf("query error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	items := make([]QueryResultItem, len(results))
	for i, r := range results {
		items[i] = QueryResultItem{
			Source:      r.SourceName,
			FileUri:     r.FileUri,
			MimeType:    r.MimeType,
			ChunkIndex:  r.ChunkIndex,
			ChunkHeader: r.ChunkHeader,
			ChunkText:   r.ChunkText,
			ExtraJSON:   r.ExtraJSON,
			Distance:    r.Distance,
			Similarity:  r.Similarity,
		}
	}

	return c.JSON(fiber.Map{"results": items})
}

func (h *Handler) Rag(c *fiber.Ctx) error {
	collection, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	var req RagRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}
	if req.TopK == 0 {
		req.TopK = 10
	}

	prompt, err := h.meta.GetPromptBySlug(c.Context(), req.PromptSlug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "prompt not found"})
	}

	emb, err := h.registry.Get(collection.EmbedModel)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no embedder available for this collection's model"})
	}

	vectors, err := emb.Embed(c.Context(), []string{req.Query})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to embed query"})
	}

	results, err := h.vectorDb.QuerySimilarVectors(c.Context(), collection.TableName, embedder.Normalize(vectors[0]), req.TopK)
	if err != nil {
		log.Printf("rag query error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	context := buildContext(results)
	formatted := strings.ReplaceAll(prompt.Content, "{{context}}", context)
	formatted = strings.ReplaceAll(formatted, "{{question}}", req.Query)

	chunks := make([]RagChunk, len(results))
	for i, r := range results {
		chunks[i] = RagChunk{
			Source:      r.SourceName,
			FileUri:     r.FileUri,
			ChunkIndex:  r.ChunkIndex,
			ChunkHeader: r.ChunkHeader,
			ChunkText:   r.ChunkText,
			Similarity:  r.Similarity,
		}
	}

	return c.JSON(RagResponse{
		FormattedPrompt: formatted,
		Chunks:          chunks,
		PromptSlug:      prompt.Slug,
	})
}

func buildContext(chunks []db.ChunkQueryResult) string {
	parts := make([]string, 0, len(chunks))
	for _, c := range chunks {
		text := ""
		if c.ChunkText != nil {
			text = *c.ChunkText
		}
		if c.ChunkHeader != nil && *c.ChunkHeader != "" {
			parts = append(parts, fmt.Sprintf("[Section: %s]\n%s", *c.ChunkHeader, text))
		} else {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}
