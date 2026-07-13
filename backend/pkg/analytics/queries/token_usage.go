package queries

import (
	"context"
	"strings"
	"time"
)

// CollectionTokens keeps all four token counters distinct rather than
// merging them into one "embed tokens" figure — ingestion-time document
// embedding and query-time embedding are different costs, worth seeing
// separately, and LLM input/output tokens are priced differently from each
// other and from embedding tokens.
type CollectionTokens struct {
	CollectionSlug       string `json:"collection_slug"`
	IngestionEmbedTokens int64  `json:"ingestion_embed_tokens"`
	QueryEmbedTokens     int64  `json:"query_embed_tokens"`
	LLMInputTokens       int64  `json:"llm_input_tokens"`
	LLMOutputTokens      int64  `json:"llm_output_tokens"`
}

const ingestionTokensSQL = `
SELECT collection_slug, COALESCE(embed_tokens, 0) AS ingestion_embed_tokens,
       0 AS query_embed_tokens, 0 AS llm_input_tokens, 0 AS llm_output_tokens
FROM ingestion_events WHERE occurred_at >= ?`

const queryTokensSQL = `
SELECT collection_slug, 0 AS ingestion_embed_tokens,
       COALESCE(embed_query_tokens, 0) AS query_embed_tokens,
       COALESCE(llm_input_tokens, 0) AS llm_input_tokens,
       COALESCE(llm_output_tokens, 0) AS llm_output_tokens
FROM query_events WHERE occurred_at >= ?`

// TokenUsageByCollection sums every token counter tracked across both
// tables (ingestion embedding, query embedding, LLM input, LLM output),
// grouped by collection_slug.
func TokenUsageByCollection(ctx context.Context, s *Store, cutoff time.Time) ([]CollectionTokens, error) {
	ingOK, err := s.ensureView(ctx, "ingestion_events")
	if err != nil {
		return nil, err
	}
	qOK, err := s.ensureView(ctx, "query_events")
	if err != nil {
		return nil, err
	}
	if !ingOK && !qOK {
		return []CollectionTokens{}, nil
	}

	var parts []string
	var args []any
	if ingOK {
		parts = append(parts, ingestionTokensSQL)
		args = append(args, cutoff)
	}
	if qOK {
		parts = append(parts, queryTokensSQL)
		args = append(args, cutoff)
	}
	stmt := `SELECT collection_slug,
			SUM(ingestion_embed_tokens) AS ingestion_embed_tokens,
			SUM(query_embed_tokens) AS query_embed_tokens,
			SUM(llm_input_tokens) AS llm_input_tokens,
			SUM(llm_output_tokens) AS llm_output_tokens
		FROM (` + strings.Join(parts, " UNION ALL ") + `) t
		GROUP BY collection_slug ORDER BY collection_slug`

	rows, err := s.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []CollectionTokens{}
	for rows.Next() {
		var c CollectionTokens
		if err := rows.Scan(&c.CollectionSlug, &c.IngestionEmbedTokens, &c.QueryEmbedTokens, &c.LLMInputTokens, &c.LLMOutputTokens); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
