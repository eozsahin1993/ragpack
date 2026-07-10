# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

RagPack: open-source, self-hosted RAG/semantic-search infrastructure for early-stage startups. A Go backend (Fiber) handles ingestion, chunking, embedding, vector search, and RAG; a Next.js admin UI manages collections/documents/queries; a Docker Compose stack (plus an `npx ragpack` CLI) is the intended way to run it.

Monorepo layout: `backend/` (Go API + engine), `web-admin/` (Next.js admin UI), `cli/` (the `ragpack` npm CLI that wraps docker compose), `sdks/` (client SDKs, e.g. `ragpack-js`), `docs/`.

## Conventions

Don't add comments that restate what the code does — only comment a genuine gotcha (a non-obvious constraint, a workaround, a "why" that isn't visible from the code itself). Prefer deriving a value/whitelist from a single source (a struct's own tags, an existing enum) over hand-copying the same names into a second list — see `pkg/meta/sort.go`'s `Sortable[T]` for the pattern.

## Commands

### Local dev (recommended — hot reload for both backend and frontend)

```bash
npm run dev           # docker compose, backend (air) + web-admin (next dev), hot reload
npm run dev:ollama    # same, plus a local Ollama container
npm run dev:tei       # same, plus a HuggingFace TEI container
```

Backend on `:9000` (public API) — admin-only routes are on a separate internal port (`:9001`) not published to the host. Admin UI on `:3000`.

Both services are bind-mounted, so edits on the host are picked up live:
- Backend: `air` (see `backend/Dockerfile.dev`) rebuilds and restarts the Go binary on every `.go` change — no manual restart needed.
- web-admin: `next dev` hot-reloads on save.

If the web-admin container's `.next` cache ever gets corrupted (stale build-manifest ENOENT errors after adding new routes while the dev server is running), fix by stopping the container, `rm -rf web-admin/.next`, and starting it again — a plain restart does not clear the bind-mounted cache.

### Backend (Go), outside Docker

`go build ./...` / `go run ./cmd` will fail on a bare host — LanceDB is used via CGO and needs its native library set up first (`CGO_CFLAGS`/`CGO_LDFLAGS` pointing at the downloaded `liblancedb_go.a`). Use `cd backend && ./dev.sh` (macOS arm64 only) instead, which downloads the artifact and sets the env for you.

`go build ./pkg/...` and `go vet ./...` work fine without any of that (they don't link the final binary), and are the fastest way to check the backend compiles/vets after an edit.

```bash
cd backend
go build ./pkg/...      # compiles all packages except cmd (no CGO link needed)
go vet ./...
go test ./...           # unit tests (pkg/parser, pkg/db/filter, pkg/util)
go test ./pkg/parser/... -run TestName   # single test
```

### Frontend (web-admin)

```bash
cd web-admin
npx tsc --noEmit        # typecheck
```

No lint or test script is configured for web-admin.

## Architecture

### Two API surfaces, one router

`pkg/api/router.go` mounts the *same* route set (`mountRoutes`) twice:
- `RegisterPublic` → `/api/v1/*`, behind `middleware.Auth` (bearer API key) — the external-facing API documented in the README, meant to be exposed to the internet.
- `RegisterAdmin` → `/admin/*`, no auth — meant only for the internal Docker network (the web-admin container talks to this), never published outside it.

Most routes are registered both top-level (cross-collection) and again under `/collections/:slug` (`middleware.Collection` resolves the slug once and both handles get reused). A handler needing collection context always resolves it from data it already has (e.g. `doc.CollectionID` on a document) rather than requiring the URL to carry `:slug` — this is what makes it safe to mount handlers like `documents.Get/Chunks/Update/Delete` at the top level too, unlike `ingest`/`query`, which genuinely need `:slug` middleware context.

### Ingestion pipeline: Job → Document → Chunks

An ingest request creates a `meta.Job` row and hands it to an in-memory worker pool (`pkg/ingester`). The worker: creates/resets the `Document` row (status `ingesting`) *before* any risky work (fetch/parse/embed), fetches (`pkg/fetcher`), parses by mime type (`pkg/parser`), chunks (`pkg/chunker`), embeds (`pkg/embedder`), then writes chunk rows to LanceDB and flips the `Document` to `complete`/`failed`. Because the Document row is created up front, almost every failure mode (bad file, embed API error, unsupported type) surfaces as `Document.status = "failed"` with an error — the only gap is a handful of pre-Document-creation lookups (bad collection ID, dedup lookup failure) which fail the job with no Document row at all.

Jobs are the internal queue/audit mechanism; end users/UI care about the resulting `Document`. Job endpoints (`/jobs`) stay in the API for debugging stuck/failed async ingests, but the admin UI's primary surface is Documents, not Jobs.

### Dual storage: SQLite (metadata) + LanceDB (vectors)

- `pkg/meta` (SQLite, via `pkg/meta/sqlite`): collections, documents, jobs, API keys, prompts, metadata field definitions — the source of truth for anything relational.
- `pkg/db/lancedb`: chunk rows (text, vector, and denormalized filter columns) — LanceDB has no native document-level table, so `extra_json` and metadata field values are *duplicated* onto every chunk row at ingest time, and re-synced on every `PATCH /documents/:id` (`UpdateChunks`/`MergeMetadataSlots`). This trades a double-write on patch for zero-latency reads on every query — worth it because reads (queries) vastly outnumber patches, and if no metadata fields are registered for a collection, queries are pure LanceDB with no join needed at all. Don't undo this duplication unless `UpdateChunks` is a measured bottleneck.

### Typed metadata fields ("slots")

A collection can register metadata fields (`pkg/meta/metadata_field.go`), each assigned a fixed-width "slot" per type (`metadata_str_1`, `metadata_num_1`, …) on the LanceDB chunk schema (`db.MetadataSlotColumn`). `pkg/ingester/metadata.go` (`MergeMetadataSlots`) maps field name → slot when writing; `pkg/db/filter` compiles user-facing filter expressions (by field name) down to slot-column predicates for queries. `GET .../documents/:id/metadata` (`consistentMetadata` in `pkg/api/documents/metadata.go`) deliberately only returns a field's value if it's identical across *every* chunk for that document — this guards against showing a value that's mid-write / partially synced.

### Generic sortable-field pattern (`pkg/meta/sort.go`)

Any store type can opt a subset of its fields into "sortable" without a hand-maintained whitelist: tag a struct field `sort:"true"` (or `sort:"default"` for the one used when no sort is requested), then call `meta.Sortable[T]()` once — it reflects over the struct's tags and returns a `SortSpec{Valid, Default}`. `pkg/api/validate.RegisterSortValidator(tag, spec)` wires that spec into a `validate:"<tag>"` struct-tag rule in one call, including a dynamic "not a valid sort field, options are [...]" error message — so a new sortable entity (e.g. Jobs, Collections) is a few lines, not a new whitelist to keep in sync by hand. See `Document`'s `sort` tags and `documents/schema.go`'s `ListQuery.SortBy` for the pattern in use.

### Embedder / LLM registries

`pkg/embedder` and `pkg/llm` each expose a `Registry` that providers (OpenAI, Ollama, HuggingFace TEI for embeddings; OpenAI, Ollama, Anthropic for LLMs) register into at startup (wired in `cmd/main.go`). A collection is pinned to one embedder + vector dimension at creation time. To add a new embedding model, edit `backend/pkg/embedder/dimensions.go`.

### Auth

Bearer API key only (`pkg/auth`, `middleware.Auth`), checked against `pkg/meta`'s API key store. On first boot with zero keys, `cmd/bootstrap.go` generates a master key, prints it once, and writes it to `<DATA_PATH>/api_key`. Admin routes (`/admin/*`) skip auth entirely — they must never be exposed outside the Docker network.
