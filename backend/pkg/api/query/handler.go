package query

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/llm"
	"ragpack/pkg/meta"
)

const (
	defaultQueryTopK = 5
	defaultRagTopK   = 2 // leaner than Query's: RAG chunks feed an LLM prompt, so cost scales with count
)

type Handler struct {
	meta              meta.MetaStore
	vectorDb          db.VectorDb
	registry          *embedder.Registry
	llmRegistry       *llm.Registry
	defaultPromptSlug string
}

func NewHandler(ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, llmRegistry *llm.Registry, defaultPromptSlug string) *Handler {
	return &Handler{meta: ms, vectorDb: vec, registry: registry, llmRegistry: llmRegistry, defaultPromptSlug: defaultPromptSlug}
}

// resolveHybridSettings fills in db.DefaultHybridSettings for any unset field.
func resolveHybridSettings(s *HybridSettings) db.HybridSettings {
	resolved := db.DefaultHybridSettings()
	if s == nil {
		return resolved
	}
	if s.FullTextWeight != nil {
		resolved.FullTextWeight = *s.FullTextWeight
	}
	if s.SemanticWeight != nil {
		resolved.SemanticWeight = *s.SemanticWeight
	}
	if s.RRFK != nil {
		resolved.RRFK = *s.RRFK
	}
	if s.FullTextLimit != nil {
		resolved.FullTextLimit = *s.FullTextLimit
	}
	return resolved
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
		req.TopK = defaultQueryTopK
	}

	metaFields, sqlFilter, err := h.resolveFilter(c.Context(), collection.ID, req.Filters)
	if err != nil {
		return writeFilterErr(c, err)
	}

	vector, err := h.embedQuery(c, collection.EmbedModel, req.Query)
	if err != nil {
		return err
	}

	keywordQuery := ""
	if !req.VectorSearchOnly {
		keywordQuery = req.Query
	}
	results, err := h.vectorDb.QuerySimilarVectors(c.Context(), collection.TableName, vector, req.TopK, sqlFilter, keywordQuery, resolveHybridSettings(req.HybridSettings))
	if err != nil {
		log.Printf("query error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	items := make([]QueryResultItem, len(results))
	for i, r := range results {
		items[i] = QueryResultItem{
			Source:             r.SourceName,
			FileUri:            r.FileUri,
			MimeType:           r.MimeType,
			ChunkIndex:         r.ChunkIndex,
			ChunkHeader:        r.ChunkHeader,
			ChunkText:          r.ChunkText,
			ExtraJSON:          r.ExtraJSON,
			VectorDistance:     r.VectorDistance,
			VectorSimilarity:   r.VectorSimilarity,
			KeywordBM25Score:   r.KeywordBM25Score,
			RRFScoreNormalized: r.RRFScoreNormalized,
			RRFScore:           r.RRFScore,
			Metadata:           reconstructMetadata(r.ChunkDbRecord, metaFields),
		}
	}

	return c.JSON(QueryResponse{Results: items})
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
		req.TopK = defaultRagTopK
	}
	if req.PromptSlug == "" {
		req.PromptSlug = h.defaultPromptSlug
	}

	prompt, err := h.meta.GetPromptBySlug(c.Context(), req.PromptSlug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "prompt not found"})
	}

	_, sqlFilter, err := h.resolveFilter(c.Context(), collection.ID, req.Filters)
	if err != nil {
		return writeFilterErr(c, err)
	}

	vector, err := h.embedQuery(c, collection.EmbedModel, req.Query)
	if err != nil {
		return err
	}

	keywordQuery := ""
	if !req.VectorSearchOnly {
		keywordQuery = req.Query
	}
	results, err := h.vectorDb.QuerySimilarVectors(c.Context(), collection.TableName, vector, req.TopK, sqlFilter, keywordQuery, resolveHybridSettings(req.HybridSettings))
	if err != nil {
		log.Printf("rag query error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if req.MinSimilarity != nil && *req.MinSimilarity > 0 {
		filtered := results[:0]
		for _, r := range results {
			if r.VectorSimilarity >= *req.MinSimilarity {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	formatted := strings.ReplaceAll(prompt.Content, "{{context}}", buildContext(results))
	formatted = strings.ReplaceAll(formatted, "{{question}}", req.Query)

	chunks := make([]RagChunk, len(results))
	for i, r := range results {
		chunks[i] = RagChunk{
			Source:             r.SourceName,
			FileUri:            r.FileUri,
			ChunkIndex:         r.ChunkIndex,
			ChunkHeader:        r.ChunkHeader,
			ChunkText:          r.ChunkText,
			VectorSimilarity:   r.VectorSimilarity,
			KeywordBM25Score:   r.KeywordBM25Score,
			RRFScoreNormalized: r.RRFScoreNormalized,
			RRFScore:           r.RRFScore,
		}
	}

	var provider llm.LLM
	if req.Model != "" {
		provider, err = h.llmRegistry.Get(req.Model)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("model %q not available: %v", req.Model, err)})
		}
	} else {
		_, provider, err = h.llmRegistry.Default()
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": fmt.Sprintf("no LLM configured: %v", err)})
		}
	}

	llmCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	answer, err := provider.Complete(llmCtx, formatted)
	if err != nil {
		log.Printf("rag llm error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "LLM call failed"})
	}

	return c.JSON(RagResponse{
		FormattedPrompt: formatted,
		Answer:          answer,
		Chunks:          chunks,
		PromptSlug:      prompt.Slug,
	})
}

func (h *Handler) embedQuery(c *fiber.Ctx, embedModel, query string) ([]float32, error) {
	emb, err := h.registry.Get(embedModel)
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no embedder available for this collection's model"})
	}
	vectors, err := emb.Embed(c.Context(), []string{query})
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to embed query"})
	}
	return embedder.Normalize(vectors[0]), nil
}

func buildContext(chunks []db.ChunkQueryResult) string {
	parts := make([]string, 0, len(chunks))
	for i, c := range chunks {
		text := ""
		if c.ChunkText != nil {
			text = *c.ChunkText
		}
		label := fmt.Sprintf("[Source %d: %s]", i+1, c.SourceName)
		if c.ChunkHeader != nil && *c.ChunkHeader != "" {
			parts = append(parts, fmt.Sprintf("%s\n[Section: %s]\n%s", label, *c.ChunkHeader, text))
		} else {
			parts = append(parts, fmt.Sprintf("%s\n%s", label, text))
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}
