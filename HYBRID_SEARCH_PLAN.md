# Hybrid Search Implementation Plan

## Overview

Add metadata pre-filtering and keyword/RRF hybrid search to ragpack's query pipeline.
Phase 1 covers metadata filters (hard requirements before search). Phase 2 covers keyword+vector hybrid search with RRF (can be done later).

## Status (as of 2026-07-10)

**Both phases are shipped.** Naming note: the commit that shipped Phase 1
was tagged `feat(hybrid-search)`, but what it built is **filtered vector
search** (embedding similarity + a structured pre-filter) — not hybrid
search in the conventional sense. "Hybrid search" properly refers to
Phase 2 — combining *two separate retrieval methods* (dense vector +
sparse keyword/BM25) via weighted RRF — which now also exists
(`vector_search_only` flag + `hybrid_settings`), see the Phase 2 section below
for implementation details.

Phase 1 also went further than this doc originally specified:
- **5 typed slot kinds shipped, not 3.** The plan below has `bool`/`date`
  string-encoded into `metadata_str_*` slots. The actual implementation
  gave `bool` and `date` their own dedicated slot arrays instead
  (`metadata_bool_1..10`, `metadata_date_1..10`), alongside `str` (bumped
  to 20 slots), `num`, and `arr` (10 slots each). Simpler filter/ingest
  code, no string coercion needed for booleans/dates.
- **Known gap**: field deletion (`DELETE /collections/:slug/metadata-fields/:name`)
  does not require `"confirm_data_deletion": true` as this doc specifies —
  it deletes immediately on request. Everything else in the deletion
  ordering (SQLite row → null slot → drop index → optimize) matches the
  plan.

---

## Phase 1: Metadata Filters

### Design Decisions

- **30 new Arrow columns** added to every chunk row (nulls are free in columnar storage):
  - `metadata_str_1..10` — utf8 nullable (text, ISO 8601 dates, booleans as "true"/"false")
  - `metadata_num_1..10` — float64 nullable (ints, floats, doubles all fit)
  - `metadata_arr_1..10` — list\<utf8\> nullable (tags, categories, multi-value fields)
- **Separate `metadata` field** at ingest time (not extracted from `extra_json`). `extra_json` stays as an opaque read-back blob.
- **Dedicated `collection_metadata_fields` SQLite table** — normalized, not a JSON column on collections. Fields must be pre-declared before ingest or query can use them. Schema:
  ```sql
  CREATE TABLE collection_metadata_fields (
      id            TEXT PRIMARY KEY,
      collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
      name          TEXT NOT NULL,
      type          TEXT NOT NULL,  -- "str" | "num" | "arr"
      slot          INTEGER NOT NULL,
      UNIQUE(collection_id, name),
      UNIQUE(collection_id, type, slot)
  );
  ```
- **Fields are declared separately** via `POST /collections/:slug/metadata-fields` — not at collection creation time. System auto-assigns the next available slot for the given type.
- **MongoDB-style filter DSL** in query requests — compiles to DataFusion SQL predicate
- **Built-in columns** (`created_at`, `mime_type`, `source_name`, `external_id`) filterable directly without using a metadata slot
- **Pre-filter semantics** — filters are hard requirements applied before ANN search via `VectorSearchWithFilter`, not scoring boosts

### Field lifecycle
- **Type changes** — treated as delete + re-insert with the new type. Not a PATCH operation. User must re-ingest affected documents to populate the new slot.
- **Field deletion** — destructive and irreversible. Requires `"confirm_data_deletion": true` in the request body. Ordering matters for concurrent safety: remove the SQLite mapping row first (so new requests stop routing to that slot), then null out the slot in LanceDB (`tbl.Update` with no filter), then drop the index, then `Optimize`. No async job needed — for ragpack's scale this completes in under a second. Lance's columnar format only rewrites the affected column's data pages.
- **Slot recycling** — safe because the slot is fully nulled before the mapping is removed. A new field can claim that slot number without risk of inheriting old data.

### Ingest behaviour with metadata
- Declared keys are routed to their slot with type coercion (bool→string, date→ISO 8601 string)
- Undeclared keys in `metadata` are silently ignored but the ingest response includes a `warnings` array listing them — makes typos debuggable without failing the ingest
- No backfill: documents ingested before a field was declared will have `NULL` in that slot and won't match filters on that field

### Date handling
- Stored as ISO 8601 strings in `metadata_str_*` slots
- Lexicographic comparison works correctly for ISO dates (`"2024-06-01" > "2024-01-01"`)
- Filter compiler should also accept relative expressions: `"last week"`, `"7 days ago"`, `"yesterday"`

### Boolean handling
- Stored as `"true"` / `"false"` strings in `metadata_str_*` slots
- Filter compiler coerces Go `true`/`false` JSON booleans to string before building SQL

### Supported filter operators
| Operator | SQL equivalent | Notes |
|---|---|---|
| `$eq` | `= 'val'` | |
| `$ne` | `!= 'val'` | |
| `$gt` | `> val` | |
| `$gte` | `>= val` | |
| `$lt` | `< val` | |
| `$lte` | `<= val` | |
| `$in` | `IN ('a','b')` | |
| `$nin` | `NOT IN ('a','b')` | |
| `$exists` | `IS NOT NULL` / `IS NULL` | |
| `$like` | `LIKE '%val%'` | case-sensitive |
| `$ilike` | `ILIKE '%val%'` | case-insensitive, verify DataFusion support |
| `$contains` | `array_has(col, 'val')` | arr slots only |
| `$containsAny` | `array_has_any(col, make_array(...))` | arr slots only |
| `$containsAll` | `array_has_all(col, make_array(...))` | arr slots only |
| `$and` | `AND` | logical |
| `$or` | `OR` | logical |

### Example filter request
```json
{
  "query": "quarterly revenue",
  "top_k": 10,
  "filters": {
    "$and": [
      {"category":   {"$eq": "finance"}},
      {"published":  {"$gt": "2024-01-01"}},
      {"price":      {"$lte": 99.99}},
      {"tags":       {"$containsAny": ["report", "earnings"]}},
      {"mime_type":  {"$eq": "application/pdf"}}
    ]
  }
}
```

---

## Build Order

### Step 1 — Arrow schema + indexes (`pkg/db/lancedb/arrow.go`, `lancedb.go`)
- Add 30 new column constants
- Extend `chunkArrowSchema(vectorDim)` to include them
- Extend `chunkBuilders`, `newChunkBuilders`, `append`, `finish`, `chunksToArrowRecord`
- Extend `rowToChunk` / `mapResultsToChunks` to read them back
- Index creation is deferred — no indexes created at `CreateTable` time. Indexes are created per-slot when a field is registered and dropped when a field is deleted. Only slots in active use carry an index.
- Index size stays proportional to what users actually use — unused slots have no index overhead.
- Add `OptimizeIndex(ctx, tableName string) error` to the `VectorDb` interface, implemented via `OptimizeWithAction(ctx, OptimizeAction{Kind: OptimizeIndex})`.
- Call `OptimizeIndex` in the ingest worker after the final `flush()` in `process()` — once per document, not per batch. Guarantees all chunks are fully indexed before the document is marked `complete`. Predictable: query performance never degrades as documents accumulate.

### Step 2 — SQLite migration
- New `collection_metadata_fields` table (schema above)
- New goose migration file in `pkg/meta/sqlite/migrations/`

### Step 3 — Meta store interface + SQLite impl
- New `MetadataField` struct: `{ID, CollectionID, Name, Type, Slot}`
- Add to `meta.MetaStore`: `RegisterMetadataField`, `ListMetadataFields`, `DeleteMetadataField`
- SQLite impl: slot auto-assignment queries per type

### Step 4 — Metadata fields API (`pkg/api/collections/`)
- `POST /collections/:slug/metadata-fields` — register fields in batch, auto-assign slots, create index per slot:
  ```json
  [
    {"name": "category", "type": "str"},
    {"name": "price",    "type": "num"},
    {"name": "tags",     "type": "arr"}
  ]
  ```
  Index type per field: `str` → `IndexTypeBitmap`, `num` → `IndexTypeBTree`, `arr` → `IndexTypeLabelList`. Entire batch runs in a single SQLite transaction (serializes concurrent registrations, prevents slot collisions). SQLite rows committed first — if any index creation then fails, compensating rollback: delete the SQLite rows and drop any already-created indexes. Can be called multiple times to add more fields later. Already-registered names are rejected with a clear error.
- `GET  /collections/:slug/metadata-fields` — list all registered fields with their assigned slots
- `DELETE /collections/:slug/metadata-fields/:name` — one field at a time (destructive); requires `"confirm_data_deletion": true`; synchronously nulls the slot in LanceDB + drops the index + optimizes + removes mapping row
- Validate: max 10 str, 10 num, 10 arr across all registered fields; no duplicate names; valid types; name not in reserved list (`created_at`, `mime_type`, `source_name`, `external_id`, `document_id`, `chunk_text`, `chunk_header`)
- Type change: rejected with clear error — "delete the field and re-register with the new type, then re-ingest"

### Step 5 — Filter compiler (`pkg/db/filter/`)
- New package: `pkg/db/filter/compiler.go`
- Input: raw JSON filter object + collection metadata mapping
- Output: DataFusion SQL predicate string
- Resolves user field names → slot column names
- Handles built-in columns directly; `created_at` converts ISO string → unix int64 before building SQL
- Compiles all operators in the table above
- Date relative expression parser (reuse `araddon/dateparse` already in go.mod)
- Filter on undeclared field name → clean API error: `"field 'genre' is not registered on this collection"`
- Type coercion failure at ingest (e.g. `"price": "not-a-number"` for a `num` field) → warning in ingest response, field skipped, ingest continues
- **SQL injection**: all user-provided filter values must be escaped before interpolation into the SQL predicate string — DataFusion is embedded so the risk is limited but sanitization is still required

### Step 6 — `VectorDb` interface (`pkg/db/store.go`)
- Add `filter string` parameter to `QuerySimilarVectors`
- Empty string = no filter (backwards compatible)

### Step 7 — LanceDB impl (`pkg/db/lancedb/lancedb.go`)
- `QuerySimilarVectors` routes to `VectorSearchWithFilter` when filter is non-empty
- Otherwise keeps `VectorSearch`

### Step 8 — Ingest pipeline (`pkg/api/ingest/`, `pkg/ingester/`)
- Add `metadata map[string]interface{}` field to ingest request schema
- Look up collection's metadata fields once at the start of `process()`, not inside `flush()` — avoids redundant SQLite reads per batch
- Route declared keys to their slots with type coercion (bool→string, date→ISO 8601 string)
- Undeclared keys: silently ignored but collected into `warnings` on the ingest response

### Step 9 — Query + RAG handlers (`pkg/api/query/`)
- Add `Filters json.RawMessage` to `QueryRequest` and `RagRequest`
- Look up collection's metadata fields (same SQLite call already needed for the collection itself)
- Parse filters, compile to SQL via filter package, pass to `QuerySimilarVectors`
- Reconstruct `metadata` object in the response: read slot columns from LanceDB result, map back to user field names using the field mapping

---

## Phase 2: Keyword + Weighted RRF Hybrid Search

**Status: shipped (2026-07-10), redesigned same day.** This is the actual
"hybrid search" feature — combining dense vector retrieval with sparse
keyword/BM25 retrieval via RRF — as distinct from Phase 1's filtered
vector search above.

The first implementation used lancedb's built-in `Reranker` to do the
fusion in a single call. That was replaced after discovering (via
`RerankerConfig`'s actual field list, not assumption) that lancedb's RRF
reranker is hard-coded 50/50 across channels with no weight knob and no
way to recover per-channel scores, so the fusion was rebuilt in-house to
be weighted and fully configurable:

- **Two independent passes, fused in Go, not lancedb.**
  `QuerySimilarVectors` (`pkg/db/lancedb/lancedb.go`) runs a plain vector
  `Select` and, when `keywordQuery` is non-empty, a separate
  `FullTextSearch`/`FullTextSearchWithFilter` pass, then merges both
  result lists via `db.MergeWeightedRRF` (`pkg/db/hybrid.go`) — backend-
  agnostic, lives in `pkg/db`, not the lancedb package.
- **Weighted RRF formula**: per channel, `weight / (rrf_k + rank)`, summed
  across whichever channel(s) a result appears in.
- **Fully configurable, per request**: `HybridSettings` (`full_text_weight`,
  `semantic_weight`, `rrf_k`, `full_text_limit`) on `QueryRequest`/
  `RagRequest`, all pointer fields so any subset can be overridden; unset
  fields fall back to `db.DefaultHybridSettings()`
  (`full_text_weight: 0.3, semantic_weight: 0.7, rrf_k: 60, full_text_limit: 200`) —
  semantic favored 7:3 by default. Only the ratio between the two weights
  matters (uniform scaling doesn't change ranking or `rrf_score_normalized`,
  only `rrf_score`'s raw magnitude).
  `full_text_limit` caps the FTS candidate pool — `FullTextSearch` has no
  native limit parameter and would otherwise return every match.
- **All scores returned, not just the fused one.** Each result carries the
  real vector cosine `similarity` (populated whenever the result was a
  vector-channel candidate), the raw BM25 `keyword_bm25_score` (populated
  whenever it was a keyword-channel candidate, unnormalized — no
  universal BM25 ceiling exists), `rrf_score` (the unclamped weighted
  fused sum, comparable only within one query's own weights/k), and
  `rrf_score_normalized` — the same value normalized against this batch's
  own top score, so the best result in any given result set is always
  100%. An earlier version normalized against a fixed ceiling derived from
  weights/k alone, which real dual-channel matches (a result found via
  both vector and keyword search) routinely exceeded — since hybrid runs
  by default and vector+FTS candidate sets overlap heavily in practice,
  that made most results clamp at 100% with no differentiation. Batch-
  relative normalization always spreads meaningfully across one query's
  own results; use `rrf_score` if you need a value comparable across
  different queries/settings.
- **`min_similarity` (RagRequest) works unconditionally now** — no special
  hybrid-mode carve-out needed, since `similarity` is always correctly
  populated for any result that came through the vector channel, hybrid
  or not. A hybrid result found only via keyword match has no vector-
  channel similarity and correctly filters out at `similarity: 0`.
- **Hybrid search runs by default.** `vector_search_only` (plain bool,
  zero-value `false`) is the opt-out — set it `true` to skip the
  keyword/FTS pass and run plain vector search. `top_k` falls back to a
  manual `if req.TopK == 0` check in the handler (same pattern as before
  this feature) — 10 for `Query()`, 2 for `Rag()` (RAG chunks get stuffed
  into an LLM prompt, so a leaner default keeps token cost down; matches
  LlamaIndex's default `similarity_top_k` of 2). Its validate tag is
  `omitempty,min=1,max=100` so an
  omitted `top_k` doesn't get rejected by validation before that default
  line runs.

---

## Verification Status

All DataFusion filter functions verified via Docker test (v1.0.4):
- ✅ `array_has`, `array_has_any`, `array_has_all`
- ✅ `ILIKE`
- ✅ `list<utf8>` column type in Arrow schema + filter
- ✅ `float64` `IS NULL` / `IS NOT NULL`
- ✅ ISO date lexicographic comparison (works correctly when slot contains only ISO dates)
- ✅ `AND`, `OR`, boolean-as-string, `LIKE`

Still to verify: whether `IndexTypeLabelList` enables indexed `array_has` or gracefully falls back to scan.

---

## Known Limitations

- 30 slots total (10 per type) is a hard ceiling; users who need more must use `extra_json` + post-filter
- Array slots are string-only (list\<utf8\>); no typed arrays
- Numeric range queries on `extra_json` fields not supported (post-filter only)
- No backfill: documents ingested before a field was declared have NULL in that slot
- Type changes require delete + re-register + re-ingest (no in-place migration)
- `array_has` on `metadata_arr_*` columns may not benefit from an index — scans at ragpack's scale are acceptable
- ~~Phase 2 hybrid search may require upgrading lancedb-go from v0.1.4 to v1.0.4 — verify before starting~~ — already on v1.0.4 (see `backend/go.mod`), this blocker is resolved
