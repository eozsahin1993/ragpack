---
sidebar_position: 4
---

# API Reference

Base URL: `http://localhost:9000/api/v1`

All requests require an API key. On first startup the backend prints it to the logs and saves it to `/data/api_key` inside the container. To retrieve it:

```bash
ragpack logs backend | grep "Key:"
```

Pass it as a bearer token in every request:

```bash
export RAGPACK_API_KEY=rp_...
```

## Collections

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections` | List all collections |
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:slug` | Get a collection |
| `DELETE` | `/collections/:slug` | Delete a collection and all its data |

```bash
curl -X POST http://localhost:9000/api/v1/collections \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Docs"}'
```

## Ingest

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/ingest` | Ingest a URL or file upload |

```bash
# Ingest a URL
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"file_uri": "https://example.com/docs/guide"}'

# Upload a file
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -F "file=@./document.pdf"
```

Supported sources: `https://`, `s3://`, file uploads.  
Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`

## Query

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/query` | Semantic or hybrid search |
| `POST` | `/collections/:slug/rag` | Retrieve chunks and generate an LLM answer |

```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query": "how do I configure authentication?", "top_k": 5}'
```

By default `query` runs **hybrid search** (vector + keyword, merged with weighted RRF). Pass `vector_search_only: true` to skip the keyword pass, `filters` for a MongoDB-style filter DSL over registered metadata fields, and `hybrid_settings` to override the RRF merge for that request:

```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "how do I configure authentication?",
    "top_k": 5,
    "filters": { "$and": [ { "category": "docs" }, { "score": { "$gte": 0.5 } } ] },
    "hybrid_settings": { "semantic_weight": 0.7, "full_text_weight": 0.3, "rrf_k": 60 }
  }'
```

Filter operators: `$and`, `$or`, `$eq`, `$ne`, `$gt`, `$gte`, `$lt`, `$lte`, `$in`, `$nin`, `$exists`, `$like`, `$ilike` (string fields), `$contains`/`$containsAny`/`$containsAll` (array fields).

Each matched chunk includes `chunk_text`, `file_uri`, `vector_distance`, and `vector_similarity` (0–100). When hybrid search ran, it also includes `keyword_bm25_score`, `rrf_score`, and `rrf_score_normalized`. `rag` returns the same chunk shape (as `chunks`) plus `answer` and `formatted_prompt`; it requires `prompt_slug` and `model`.

## Documents

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections/:slug/documents` | List ingested documents (paginated) |
| `GET` | `/collections/:slug/documents/:id` | Get a document |
| `GET` | `/collections/:slug/documents/:id/metadata` | Get a document's typed metadata field values |
| `PATCH` | `/collections/:slug/documents/:id` | Update `name`, `extra_json`, and/or `metadata` |
| `DELETE` | `/collections/:slug/documents/:id` | Delete a document and its chunks |
| `GET` | `/collections/:slug/documents/:id/chunks` | List all chunks |

`PATCH` accepts any combination of `name`, `extra_json` (a JSON string), and `metadata` (a map merged into the collection's registered typed metadata fields — unregistered or mistyped keys are silently dropped):

```bash
curl -X PATCH http://localhost:9000/api/v1/collections/my-docs/documents/doc_123 \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "Q3 Report (final)", "metadata": {"status": "published"}}'
```

Every document route above is also mounted at the top level (`/documents/:id`, `/documents/:id/chunks`, `/documents/:id/metadata`) without a collection slug, resolving the collection from the document itself.

## Jobs

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections/:slug/jobs` | List ingestion jobs (also at top-level `/jobs`) |
| `GET` | `/collections/:slug/jobs/:id` | Get a job |
| `DELETE` | `/collections/:slug/jobs/:id` | Delete a job |

Jobs are the internal ingestion queue/audit trail — for tracking a document's ingest status, prefer `GET /documents/:id` (`status`: `ingesting`/`complete`/`failed`). Jobs are mainly useful for debugging stuck or failed async ingests.

## Metadata fields

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/metadata-fields` | Register a typed metadata field |
| `GET` | `/collections/:slug/metadata-fields` | List registered metadata fields |
| `DELETE` | `/collections/:slug/metadata-fields/:name` | Remove a metadata field |

## API keys

| Method | Path | Description |
|---|---|---|
| `GET` | `/keys` | List API keys |
| `POST` | `/keys` | Create an API key |
| `DELETE` | `/keys/:id` | Delete an API key |

Keys carry two independent grant kinds, both required atomically at creation and both fail-closed (no grants = no access):

- **`grants`** — per-collection access: `{"collection_slug": "my-docs", "permission": "read" | "write" | "both"}`. Omit `collection_slug` for a wildcard grant covering every collection, including ones created later.
- **`admin_grants`** — instance-administration access, gated separately from collection content: `{"resource_type": "keys" | "prompts" | "collections" | "*", "permission": "read" | "write" | "both"}`.

```bash
# Scoped collection key
curl -X POST http://localhost:9000/api/v1/keys \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "ingestion-bot", "grants": [{"collection_slug": "my-docs", "permission": "write"}]}'

# Admin-only key (manages keys and prompts, no collection access)
curl -X POST http://localhost:9000/api/v1/keys \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "ops-admin", "admin_grants": [{"resource_type": "keys", "permission": "both"}, {"resource_type": "prompts", "permission": "both"}]}'
```

The plaintext key is only ever returned once, in the `POST` response.

## Prompts

| Method | Path | Description |
|---|---|---|
| `GET` | `/prompts` | List RAG prompt templates |
| `POST` | `/prompts` | Create a prompt template |
| `GET` | `/prompts/:slug` | Get a prompt template |
| `PATCH` | `/prompts/:slug` | Update a prompt template |
| `DELETE` | `/prompts/:slug` | Delete a prompt template |

## Embedders and LLMs

| Method | Path | Description |
|---|---|---|
| `GET` | `/embeddings` | List registered embedding models and the default |
| `GET` | `/llms` | List registered LLM models and the default |
