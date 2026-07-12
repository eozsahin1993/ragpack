---
sidebar_position: 4
---

# Query

RagPack has two query modes: **semantic search**, which returns ranked chunks, and **RAG**, which retrieves chunks and uses an LLM to produce a grounded answer. Both run **hybrid search** by default — a vector pass and a keyword/BM25 pass, fused server-side via weighted RRF (semantic-favored 7:3).

## Semantic search

`findSimilar` embeds your query and returns the most relevant chunks from the collection, ranked by similarity score.

```ts
const results = await collection.findSimilar({
  query: "how do I configure authentication?",
  topK: 5,
});

for (const r of results) {
  console.log(r.vector_similarity, r.chunk_text);
}
```

| Option | Type | Description |
|---|---|---|
| `query` | `string` | Natural language search query |
| `topK` | `number` | Number of results to return. Defaults to `5`, max `100`. |
| `filters` | `FilterExpression` | Restrict results to chunks whose document matches this filter expression |
| `vectorSearchOnly` | `boolean` | Skip the keyword/BM25 pass and use pure vector search. Hybrid search runs by default. |
| `hybridSettings` | `HybridSettings` | Per-request override of the weighted RRF merge |

Each result contains:

| Field | Type | Description |
|---|---|---|
| `source` | `string` | Display name of the source document |
| `chunk_text` | `string \| null` | The matched text |
| `chunk_header` | `string \| null` | Section/heading breadcrumb the chunk was split under, if any |
| `file_uri` | `string` | Source document URI |
| `mime_type` | `string` | MIME type of the source document |
| `chunk_index` | `number` | Position of this chunk within the document |
| `extra_json` | `string \| null` | Freeform metadata attached at ingest time |
| `metadata` | `Record<string, unknown> \| undefined` | Typed metadata field values registered for this collection |
| `vector_distance` | `number` | Raw vector distance. Lower is more similar. |
| `vector_similarity` | `number` | Cosine similarity score (0–100). Higher is more relevant. |
| `keyword_bm25_score` | `number \| undefined` | Raw BM25 score from the keyword channel; present only for hybrid results |
| `rrf_score` | `number \| undefined` | Raw RRF fusion score; only comparable within this query's own weights/k; hybrid only |
| `rrf_score_normalized` | `number \| undefined` | RRF fusion score normalized so this batch's top result is 100; hybrid only |

Response fields are returned exactly as the API sends them (snake_case), while request options use camelCase — the SDK converts for you.

## Filtering by metadata

`filters` is a MongoDB-style expression compiled server-side into a predicate over a document's built-in (`mime_type`, `source_name`, `file_uri`, `created_at`, `updated_at`, …) and registered metadata fields.

```ts
await collection.findSimilar({
  query: "refund policy",
  filters: {
    $and: [
      { mime_type: "application/pdf" },
      { $or: [{ tags: { $contains: "billing" } }, { created_at: { $gte: "7 days ago" } }] },
    ],
  },
});
```

Supported operators: `$eq`, `$ne`, `$gt`, `$gte`, `$lt`, `$lte`, `$in`, `$nin`, `$exists`, `$like`/`$ilike` (string fields), `$contains`/`$containsAny`/`$containsAll` (array fields), plus `$and`/`$or` for nesting.

## Hybrid search tuning

Override the RRF merge for a single request with `hybridSettings`, or skip the keyword pass entirely with `vectorSearchOnly`:

```ts
const results = await collection.findSimilar({
  query: "password reset",
  vectorSearchOnly: false,
  hybridSettings: {
    semanticWeight: 0.7,
    fullTextWeight: 0.3,
    rrfK: 60,
    fullTextLimit: 200,
  },
});
```

## RAG (retrieve + generate)

`rag` runs the full pipeline server-side: it retrieves the most relevant chunks, builds context, fills in your prompt template, and calls the configured LLM — returning both the answer and the chunks used to produce it.

```ts
const { answer, chunks } = await collection.rag({
  query: "How do I reset my password?",
  topK: 5,
  promptSlug: "basic-rag",
  model: "gpt-4o",
  minSimilarity: 60,
});

console.log(answer);
```

| Option | Type | Description |
|---|---|---|
| `query` | `string` | The user's question |
| `topK` | `number` | Number of chunks to retrieve. Defaults to `2`, since RAG chunks feed an LLM prompt. |
| `promptSlug` | `string` | Slug of the prompt template to use. Defaults to `"basic_rag"`. |
| `model` | `string` | LLM model name (e.g. `"gpt-4o"`, `"llama3"`). Falls back to server default if omitted. |
| `minSimilarity` | `number` | Minimum similarity score (0–100) a chunk must meet to be included. Omit to include all top-K results. |
| `filters` | `FilterExpression` | Restrict retrieval to chunks whose document matches this filter expression |
| `vectorSearchOnly` | `boolean` | Skip the keyword/BM25 pass and use pure vector search. Hybrid search runs by default. |
| `hybridSettings` | `HybridSettings` | Per-request override of the weighted RRF merge |

The response contains:

| Field | Type | Description |
|---|---|---|
| `answer` | `string` | LLM-generated answer |
| `chunks` | `RagChunk[]` | Chunks used to build the context: `source`, `file_uri`, `chunk_index`, `chunk_header`, `chunk_text`, `vector_similarity`, plus `keyword_bm25_score`/`rrf_score`/`rrf_score_normalized` when hybrid ran. No `mime_type`, `extra_json`, `metadata`, or `vector_distance` — those are `findSimilar`-only. |
| `formatted_prompt` | `string` | The fully expanded prompt sent to the LLM |
| `prompt_slug` | `string` | The prompt template used |
