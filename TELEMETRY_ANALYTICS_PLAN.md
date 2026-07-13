# Telemetry & Analytics Implementation Plan

## Decisions (2026-07-12) — supersede conflicting sections below

1. **No OTel SDK.** The event model is flat fact tables with nested Parquet columns; OTel span attributes only carry flat primitives, so riding on spans would mean JSON-encoding structs into string attributes and decoding them back in the exporter — while using none of OTel's actual features (no Collector, single process, no cross-service hops). `pkg/telemetry` is a direct typed-event recorder instead: handlers fill event structs → buffered channel → background goroutine → Parquet. `Record*()` is the seam where OTel could be slotted in later without touching call sites, if RagPack ever needs to emit traces into a user's observability stack.
2. **`query_review` cut from v1.** Correctness checking = an admin eyeballing the `query_traces` drill-down (query, chunks, prompt, answer side by side). Recording verdicts (`hallucinated` etc.) is a labeling workflow that can be added later as an isolated table + one admin endpoint if tallies are ever wanted. The verdict-taxonomy open question is thereby moot.
3. **All aggregation happens at read time** — DuckDB SQL directly over the Parquet files, per dashboard load. No ETL/rollup layer in v1; pre-aggregated rollup files are the known later add-on (hourly/nightly job writing more Parquet) and require no upstream changes.
4. **Field governance:** the recorder stamps `event_id`/`occurred_at` centrally when producers leave them empty. Exception: a trace's `event_id` is a foreign key to its query event — never auto-generated; traces without one are dropped.
5. **Recording is fire-and-forget:** bounded channel, drop-on-full (the RAG pipeline never blocks on telemetry); flush every 60s or 500 rows per table into `<DATA_PATH>/telemetry/<table>/<YYYY-MM-DD>/<ts>.parquet`; `TELEMETRY_ENABLED` kill switch (default on); admin-surface calls get null `api_key_id` + `origin` column (`public`/`admin`).

## Overview

Instrument the ingestion and query/RAG pipelines to power an admin-only analytics dashboard covering: general usage (volume over time), performance (latency by stage), cost (tokens/$ for both embedding and LLM calls), and — the harder one — a way to manually check whether RAG queries returned the correct chunks and didn't hallucinate. That's the entire goal; nothing here is meant to be a public/API-facing feature. Storage/query architecture (OTel SDK spans → custom exporter → local Parquet → DuckDB, no OTel Collector, S3 as the future scale-out target) was already decided — see memory `project_telemetry_architecture`. This doc covers the part that wasn't designed yet: **what gets captured, from where, and in what shape.**

**Trust boundary, stated explicitly:** everything in this doc is internal/admin-side. Instrumentation observes requests the public API already handles — it never adds fields to `QueryResponse`/`RagResponse` or any other public response, never adds a public endpoint, and requires zero integration work from whatever application is calling RagPack's `/api/v1/*` API. The dashboard, the trace drill-down, and the `query_review` annotation endpoint all live under `/admin/*` (no auth, internal Docker network only) — same boundary as the rest of the admin surface today.

## Gaps in the current design

These are blockers or blind spots found while reading the actual pipeline code (`pkg/ingester/worker.go`, `pkg/api/query/handler.go`, `pkg/llm/*`, `pkg/embedder/*`), not hypothetical concerns:

1. **No token usage anywhere.** `LLM.Complete(ctx, prompt) (string, error)` and `Embedder.Embed(ctx, texts) ([][]float32, error)` return no usage data. OpenAI's and Anthropic's HTTP responses both include a `usage` object (`prompt_tokens`/`completion_tokens`, `input_tokens`/`output_tokens`) that today is parsed past and discarded (`pkg/llm/openai.go:70-79`, `pkg/llm/anthropic.go:63-68`). **This has to be fixed before any cost/token telemetry is possible** — it's an interface change across 3 LLM providers and 4 embedder providers, not a telemetry-layer concern.
2. **Tokens ≠ cost.** You need a pricing table (input/output $ per token, model-specific, changes over time). Local models (Ollama) genuinely cost $0 — that must be distinguishable from "unknown model, unpriced" (null), or dashboards will silently under-report spend for any model someone forgets to add to the table.
3. **Zero timing instrumentation exists today** in either pipeline. Every span boundary below is new.
4. **No collection dimension on anything.** Every event must carry `collection_id`/`collection_slug` — "which collection is costing the most" is a near-certain first dashboard question, and there's currently no telemetry data to slice by collection at all.
5. **No caller/API-key dimension.** `pkg/auth` + the new API Keys feature already give you a caller identity per request — thread the (non-secret) API key ID onto query/RAG events now, or per-consumer cost breakdown is unrecoverable later without a data backfill.
6. **Failure paths currently just `log.Printf` and return** (`worker.go:79`, `handler.go:87,151,200`). If spans are only emitted on the happy path, analytics will systematically undercount the most useful signal — which documents fail and why, what fraction of queries error.
7. **Rate-limiter wait is invisible.** `wp.limiter.Wait(ctx)` in `worker.go:186` can dominate ingestion latency under load; if it's not measured separately from the actual embed call, slow ingestion gets misattributed to "the embed API is slow" when it's actually your own throttling.
8. **Hybrid search settings aren't captured.** `FullTextWeight`/`SemanticWeight`/`RRFK` (`handler.go:37-55`) materially affect result quality and latency — needed as attributes if you want to analyze "different queries, their performance."
9. **Storing raw query text is a deliberate choice, not a default.** You want "different questions executed" — that means persisting `req.Query` verbatim. That's the user's own usage data (not chunk content), so it's a reasonable call for a self-hosted tool, but flag it as intentional rather than an oversight, since it's the one field in this whole design that could contain something sensitive.

## Event model

Two flat fact tables, one row per pipeline invocation — not raw nested OTel spans (see prior discussion: OTLP-JSON's nested shape is exactly what made querying awkward; a flat/lightly-nested Parquet schema you control avoids that).

### `ingestion_events` — one row per document per job attempt

| Column | Type | Notes |
|---|---|---|
| `event_id` | string (uuid) | |
| `occurred_at` | timestamp | |
| `job_id`, `document_id` | string | ties back to SQLite `jobs`/`documents` for drill-down |
| `collection_id`, `collection_slug` | string | |
| `file_uri`, `mime_type` | string | |
| `intent` | string | `ingest` \| `refresh` |
| `status` | string | `complete` \| `failed` |
| `error` | string, nullable | |
| `chunk_count` | int | |
| `fetch_ms` | int | around `fetcher.New`+`Fetch`, `worker.go:102-109` — 0 when a reader was pre-supplied (upload path) |
| `parse_chunk_ms` | int | see note below — not independently measurable per stage |
| `rate_limit_wait_ms` | int | sum of `wp.limiter.Wait` calls, `worker.go:186` |
| `embed_ms` | int | sum of `emb.Embed` calls, `worker.go:189` |
| `insert_ms` | int | sum of `vectorDb.InsertBatch` calls, `worker.go:217` |
| `optimize_index_ms` | int | `worker.go:241` |
| `total_ms` | int | full `process()` wall time |
| `embed_model` | string | |
| `embed_tokens` | int, nullable | from the new `Embedder` usage return |
| `embed_cost_usd` | float, nullable | null = unpriced model, not zero |

**Gotcha on stage timing:** `worker.go:225-238` fuses parse → chunk → batch → embed → insert into one streaming `for chunk, err := range chunker.Chunk(parser.Parse(ctx, reader))` loop specifically so only `batchSize` chunks are ever in memory. That means `parse` and `chunk` can't be cleanly separated from each other without changing the parser/chunker interfaces to self-report timing — not worth it for this. The clean seam is around `flush()`, whose three internal calls (`limiter.Wait`, `emb.Embed`, `vectorDb.InsertBatch`) are already three distinct, independently-timeable calls. So: time those three directly, and derive `parse_chunk_ms = loop_wall_time − (rate_limit_wait_ms + embed_ms + insert_ms)` rather than inventing four independent named stages that don't exist in the code.

### `query_events` — one row per `Query` or `Rag` call

| Column | Type | Notes |
|---|---|---|
| `event_id`, `occurred_at` | | |
| `collection_id`, `collection_slug` | string | |
| `api_key_id` | string, nullable | caller identity, not the raw key |
| `endpoint` | string | `query` \| `rag` |
| `query_text` | string | see gap #9 |
| `top_k`, `vector_search_only` | | |
| `hybrid_settings` | STRUCT\<full_text_weight, semantic_weight, rrf_k, full_text_limit\> | Parquet's nested types fit this naturally — no need to flatten into 4 separate columns |
| `filters_json` | string, nullable | the metadata filter DSL used, if any |
| `embed_model`, `embed_query_tokens` | | query embedding is one `Embed` call, `handler.go:212-222` |
| `vector_search_ms` | int | around `vectorDb.QuerySimilarVectors`, `handler.go:85` / `149` |
| `embed_ms` | int | around `embedQuery` |
| `result_count` | int | |
| `results` | LIST\<STRUCT\<source_name, similarity, bm25_score, rrf_score\>\> | top-k as a native nested column — DuckDB `UNNEST`s this fine when you want per-result rows |
| `status`, `error` | | |
| `total_ms` | int | |
| *(rag only)* `prompt_slug`, `llm_model` | | |
| *(rag only)* `llm_input_tokens`, `llm_output_tokens` | int, nullable | from the new `LLM` usage return |
| *(rag only)* `llm_cost_usd` | float, nullable | |
| *(rag only)* `llm_ms` | int | around `provider.Complete`, `handler.go:198` |

Both handlers (`Query`, `Rag`) are already linear/sequential (no streaming-generator complication like the ingester), so wrapping each call with a timer is straightforward. One implementation constraint: the telemetry wrapper must only *observe* — several error branches in `handler.go` (e.g. `embedQuery`, lines 213-221) already write the Fiber response as part of their return value, so instrumentation code must never call `c.Status`/`c.JSON` itself, only read duration/error and let the existing return value pass through untouched.

`event_id` is generated and stored server-side only — it is **not** added to `QueryResponse`/`RagResponse` or exposed to the API caller in any way. An admin browsing the dashboard finds a call by collection/time/query-text filters, not by an ID a caller would need to supply; the drill-down and review flows below operate entirely inside `/admin/*`.

### `query_traces` — heavy content, one row per query/RAG call, joined to `query_events` by `event_id`

`query_events` is deliberately lean (metric fields only) so routine dashboard aggregation (`COUNT`, `AVG(cost)`, time-bucketed volume) stays fast over a table that's scanned constantly. The "see what chunks it got" drill-down needs the actual text, which is a different, much rarer access pattern (one admin, one click, one row) — so it belongs in a sibling table, not bolted onto `query_events`:

| Column | Type | Notes |
|---|---|---|
| `event_id` | string | FK to `query_events.event_id` |
| `chunks` | LIST\<STRUCT\<source_name, chunk_text, chunk_header, similarity, bm25_score\>\> | the actual retrieved text, not just scores |
| `formatted_prompt` | string, nullable | rag only — the prompt actually sent to the LLM, `handler.go:165-166` |
| `answer` | string, nullable | rag only — `handler.go:198` |

This keeps the "no heavy raw text in spans" instinct from Phase 1 intact for the high-frequency ingestion side (hundreds of chunk-embed events per document), while still giving RAG calls — much lower volume, one row per end-user question — full traceability when an admin actually opens one.

### `query_review` — admin correctness annotation, append-only, one row per review

The goal here (per your clarification) is: general usage/perf/cost dashboard, plus a way to manually check whether the right chunks came back and the answer isn't hallucinated. That's **an admin reviewing traces inside the dashboard**, not end-users submitting feedback through the public API — so this is admin-only surface, same trust boundary as everything else under `/admin/*` (no auth, internal Docker network only, per the existing pattern). No public endpoint, no requirement that calling applications integrate anything.

There's no rating/review concept anywhere in the codebase today (confirmed by grep) — this is new surface area:

| Column | Type | Notes |
|---|---|---|
| `event_id` | string | FK to `query_events.event_id` — the same `query_id` used for the trace drill-down |
| `reviewed_at` | timestamp | |
| `verdict` | string | `correct` \| `wrong_chunks` \| `hallucinated` \| `other` — a small fixed taxonomy is more useful here than a blind thumbs up/down, since the actual goal is diagnosing *why* something's wrong, not just that it is |
| `note` | string, nullable | free-text, e.g. "should have matched doc X but didn't" |

New admin-only endpoint: `POST /admin/query-events/:event_id/review`, body `{verdict, note?}`, written by the web-admin trace view. Reviews arrive asynchronously, well after the originating `query_events`/`query_traces` rows were flushed to an immutable Parquet file — so this is its own append-only event stream (same exporter mechanism as everything else), never an attempted in-place update of an already-written row. Joined at query time: `query_events LEFT JOIN query_review USING (event_id)`.

### Full-journey drill-down

Given a `query_id`: `query_events` row (metrics, status, cost, latency) + `query_traces` row (the actual retrieved chunks, the formatted prompt, and the LLM's answer — everything needed to judge grounding) + `query_review` row if one exists (an admin's verdict). Three joins on one key, no denormalization needed — this is the "click into one query, see the chunks and answer side by side, judge whether it hallucinated" view, and it's the whole point of `query_traces` existing separately from the lean metrics table.

### Common dashboard questions

| Question | Table(s) | Sketch |
|---|---|---|
| What did re-ingesting document X cost? | `ingestion_events` | `SUM(embed_cost_usd) WHERE document_id = ? AND intent = 'refresh'` |
| Total spend this week, by collection | `ingestion_events` + `query_events` | `SUM(cost) GROUP BY collection_slug, date_trunc('week', occurred_at)` |
| RAG/query volume per hour | `query_events` | `COUNT(*) GROUP BY date_trunc('hour', occurred_at), endpoint` |
| Most expensive collection to run | `ingestion_events` + `query_events` | `SUM(embed_cost_usd + llm_cost_usd) GROUP BY collection_slug ORDER BY 1 DESC` |
| p95 RAG latency | `query_events` | `quantile_cont(total_ms, 0.95) WHERE endpoint = 'rag'` |
| Most common questions asked | `query_events` | `GROUP BY query_text ORDER BY COUNT(*) DESC` (exact-match grouping to start; semantic clustering of near-duplicate phrasings is a later problem, not v1) |
| % of reviewed RAG answers marked hallucinated/wrong chunks | `query_events` + `query_review` | `AVG(verdict != 'correct') WHERE verdict IS NOT NULL` |
| Which answers need prompt/chunking tuning | `query_events` + `query_traces` + `query_review` | `WHERE verdict IN ('wrong_chunks','hallucinated')` then open the joined trace |
| Ingestion failure rate by mime type | `ingestion_events` | `AVG(status = 'failed') GROUP BY mime_type` |
| Are we getting throttled on ingest? | `ingestion_events` | trend `rate_limit_wait_ms` over time — rising trend means the rate limiter, not the provider, is the bottleneck |

## Required interface changes (prerequisite, not part of the telemetry layer itself)

```go
// pkg/llm/llm.go
type Usage struct {
    InputTokens  int
    OutputTokens int
}

type LLM interface {
    Complete(ctx context.Context, prompt string) (string, Usage, error)
    Model() string
}
```

Each provider's `Complete` needs its usage struct decoded from a response field it currently ignores:
- OpenAI: `result.Usage.PromptTokens` / `CompletionTokens` (add `Usage` to the anonymous struct in `openai.go:70-76`)
- Anthropic: `result.Usage.InputTokens` / `OutputTokens` (add to `anthropic.go:63-67`)
- Ollama: no cost, but `prompt_eval_count`/`eval_count` are in its response and map onto the same `Usage` fields — useful for local throughput analysis even at $0 cost

Same shape for `Embedder`:

```go
// pkg/embedder/embedder.go
type Usage struct {
    TotalTokens int
}

type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, Usage, error)
    Dimensions() (int, error)
    Model() string
}
```

OpenAI's embeddings response has `usage.total_tokens` (`openai.go:83-90`, currently discarded). Ollama/HuggingFace/TEI may not report it — return a zero-value `Usage{}` and treat `embed_tokens` as null downstream, not zero, so it's visually distinct from "confirmed zero-cost."

Every call site of `Embed`/`Complete` (`worker.go:189`, `handler.go:198,217`, `probeDimensions` in `embedder.go`) needs its call signature updated to take the new return value — mechanical but touches several files.

## Pricing tables

New small lookup, same spirit as the model-registration pattern already used for provider config:

```go
// pkg/llm/pricing.go
type Pricing struct {
    InputPerMTok  float64 // USD per 1M input tokens
    OutputPerMTok float64
}

var pricingTable = map[string]Pricing{
    "gpt-4o": {InputPerMTok: 2.50, OutputPerMTok: 10.00},
    // ... one entry per priced model; local/Ollama models intentionally absent
}

func Cost(model string, u Usage) (usd float64, priced bool) { ... }
```

Same idea for embedding models (`pkg/embedder/pricing.go`), single $/1M-token rate instead of input/output split. `priced bool` is what lets `embed_cost_usd`/`llm_cost_usd` be null (unpriced/unknown model) instead of silently `0` (which reads identically to "confirmed free" on a dashboard).

## Instrumentation layer

A `pkg/telemetry` package wraps the OTel Go SDK so `worker.go` and `handler.go` don't hand-roll span/attribute code inline:

```go
span, end := telemetry.StartIngestSpan(ctx, collection, job, document)
defer end(&err) // records status/error, emits the ingestion_events row on span end
```

The custom `SpanExporter` (per the earlier architecture decision) is where flattening OTel spans down into the two Parquet schemas above actually happens — `worker.go`/`handler.go` only need to set well-known attribute keys (`collection.id`, `embed.tokens`, `llm.cost_usd`, …), not know anything about Parquet.

## Retention & size caps

- Default retention: **14 days**, configurable via `TELEMETRY_RETENTION_DAYS` (env var, same convention as the rest of `config.Config`).
- Size cap **on by default**, not opt-in: `TELEMETRY_MAX_SIZE_MB` defaults to something like 500MB, configurable/raisable. RagPack targets startups across a wide range of traffic — from near-zero to a team that's found real usage — and there's no way to know in advance which one a given self-hosted deployment is. Retention alone assumes volume stays modest enough that 14 days never gets huge; the size cap is what actually holds regardless of whether that assumption turns out true, so it shouldn't require an admin to notice they need it only after their disk fills up.
- Enforcement: a janitor goroutine on a ticker, same shape as `pkg/ingester/ingester.go`'s `loop()` (`time.NewTicker`, `ingester.go:76`) — runs hourly, lists the Parquet files under all four tables, deletes anything past the retention window, then prunes oldest-first against the size cap.
- **Granularity is per-file, not per-row.** Parquet files are immutable, so pruning can only delete whole rotated files — never trim old rows out of a file that also holds recent ones. At dev-scale rotation (small, frequent files) this makes "14 days" mean "at least 14 days," which is fine.
- **One interaction worth flagging**: `query_review` rows point at `query_events`/`query_traces` by `event_id`. If the underlying event/trace gets pruned before an admin reviews it, the trace becomes unopenable — a verdict with nothing left to look at. Recommend just accepting this for now (14 days is a generous window for someone actively using the dashboard) rather than building exemption logic for reviewed-but-not-yet-reviewed events; revisit only if this actually bites someone.

## Tension with RagPack's own goals — read before building

RagPack's pitch is self-hosted, minimal-footprint, "up and running in minutes" for early-stage startups (see memory `project_ragpack_vision`). Embedded DuckDB is in real tension with that, worth weighing before writing code, not after:

1. **A second CGO native dependency — smaller cost than it first looks, because of how LanceDB's is already handled.** Checked `backend/Dockerfile.base`/`Dockerfile`/`dev.sh`: the LanceDB native-lib download only happens in `Dockerfile.base`, a builder image published to `ghcr.io/eozsahin1993/ragpack-builder` and rebuilt only "when upgrading Go or lancedb-go versions" (its own comment). The actual prod `Dockerfile` just does `FROM ghcr.io/.../ragpack-builder:latest` — no network fetch at all, the lib is already sitting in that cached base image. Only `dev.sh` (macOS-arm64-only, for contributors building the Go backend outside Docker) downloads at first run, and that's cached after. **So a typical `docker compose up`/`npx ragpack` user never pays this cost — the maintainer does, occasionally, on version bumps.** The right move for DuckDB is the same shape of change: one more `RUN` step in `Dockerfile.base` fetching DuckDB's native lib, genuinely parallelizable with the existing LanceDB fetch since they're independent downloads (`curl duckdb & curl lancedb & wait`, or two parallel cached `RUN` layers). What's left after that correction is real but smaller: `Dockerfile.base` now owns two native artifacts to keep pinned/in-sync instead of one (a maintainer-facing cost, at version-bump time), and the final image/binary is somewhat larger (both libs linked into `/ragpack`).
2. **Blast-radius inversion.** This is meant to be a secondary, admin-only feature, but it shares a process (and container memory limit) with the actual RAG-serving API. An unbounded or accidental analytics query — the exact kind of ad hoc query an admin exploring cost data would write — can OOM the whole container and take down ingestion/query serving, the thing RagPack actually exists to do, on behalf of a "nice to have" dashboard. **Decided**: cap this at the DuckDB connection level —
   - `PRAGMA memory_limit='256MB'` (configurable via `TELEMETRY_DUCKDB_MEMORY_LIMIT`) and `PRAGMA threads=2` (configurable via `TELEMETRY_DUCKDB_MAX_THREADS`) — both GLOBAL DuckDB settings (confirmed via spike: shared across every pooled connection, not multiplied per connection), so this is one hard ceiling for the whole engine regardless of how many connections are open, not a budget that grows with concurrency.
   - A context deadline (e.g. 10s) on the admin analytics query endpoint, so a slow query degrades/cancels instead of hanging or piling up.
   - These caps don't remove the shared-fate problem (both features still live in one process/one memory limit) — they only bound how bad it can get. **Decided fallback**: if the caps ever turn out not to be enough in practice, split DuckDB out into its own process at that point — not something to build preemptively now, just the known escape hatch if this bites.
3. ~~Is a full embedded OLAP engine even earning its keep at this volume?~~ **Decided: ship it by default, not opt-in.** Given point 1's correction — the image is already Ubuntu 24.04 + poppler-utils + LanceDB linked in, not a stripped-down alpine/scratch build chasing single-digit-MB size — one more statically-linked analytical engine is a marginal percentage increase, not a category change. Not worth the added complexity of a build tag or a separate optional container just to gate something this small. The blast-radius mitigation in point 2 (memory/thread caps, query timeout, and separate-process as a later fallback) is the actual safeguard here, not making the feature optional.
4. **A silent ceiling on the one direction these startups are supposed to grow in.** Local-file embedded DuckDB can't span replicas. The day someone scales the backend horizontally — before ever doing the planned S3 migration — the dashboard won't error, it'll just quietly show partial data from whichever replica happens to hold the files on disk. Worse than a crash: the core product keeps working while the tool you're using to judge its cost/correctness goes silently wrong. **Note this has a different fix than point 2's problem** — process-splitting doesn't help here; the actual fix is the S3 migration already planned, which is what makes the data visible/consistent across replicas in the first place.

Net: this ships as a default part of RagPack, not gated behind a flag. The packaging/footprint concern turned out smaller than it first looked (point 1), the shared-memory blast radius is mitigated with caps now and process-splitting is the fallback if that's ever not enough (point 2), and the multi-replica ceiling (point 4) is already covered by the planned S3 migration, not a new problem to solve here.

## Rollout phases

1. **Interface changes** — `Usage` on `LLM`/`Embedder`, pricing tables, update all provider implementations and call sites. No telemetry yet; this alone unblocks everything else.
2. **Instrumentation** — `pkg/telemetry`, wrap `worker.go`'s `process()` and `query/handler.go`'s `Query`/`Rag`, custom exporter writing local Parquet.
3. **Query surface** — DuckDB wired into an admin-only endpoint (fits the existing "admin routes never leave the Docker network" pattern) + a web-admin analytics page.
4. **Scale-out** — swap local Parquet writes for S3, per the earlier decision; no schema or query-layer changes needed.

## Text redaction

Default is **store raw** — `query_text` in `query_events`, and chunk/prompt/answer text in `query_traces`. This isn't just a default, it's close to a requirement: the entire point of this system is letting an admin look at a query, the chunks it retrieved, and the answer, to judge correctness/hallucination — redact the text and there's nothing left to review.

Still worth a config escape hatch for anyone with compliance constraints: `TELEMETRY_REDACT_TEXT` (env flag, default `false`). When set, skip writing `query_text` and the `query_traces` table entirely rather than writing garbled/partial data — and this should be documented as "disables the correctness-review feature," not as a selective redaction, so nobody enables it expecting the drill-down to still work.
- **Verdict taxonomy**: is `correct` / `wrong_chunks` / `hallucinated` / `other` the right set, or do you want something finer-grained (e.g. separating "hallucinated" from "chunks were right but the answer still misused them")?
