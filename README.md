# RagPack

Self-hostable semantic search and RAG infrastructure for developers. Bring your own AI — Ollama for local models or OpenAI — and be up and running in minutes.

## What it does

- Ingest documents from URLs, file uploads, or S3 — RagPack fetches, parses, chunks, and embeds them automatically
- Organize documents into **collections** and query them with natural language
- Get back ranked chunks with similarity scores, ready to drop into any LLM prompt
- Manage everything via REST API or the built-in admin UI

Supported formats: `.txt`, `.md`, `.html`, `.pdf`

## Quick start

```bash
npx ragpack init       # creates .env.ragpack in the current directory
npx ragpack start      # starts the stack (API on :9000, admin UI on :3000)
```

With Ollama (fully local):

```bash
npx ragpack start --profile ollama
```

With OpenAI — edit `.env.ragpack` first:

```env
EMBED_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-3-small
```

Then `npx ragpack start`.

Open [http://localhost:3000](http://localhost:3000) for the admin UI.

## CLI

| Command | Description |
|---|---|
| `ragpack init` | Create `.env.ragpack` in the current directory |
| `ragpack start [--profile ollama]` | Start the stack |
| `ragpack stop [-v]` | Stop the stack (`-v` removes volumes and all data) |
| `ragpack logs [service]` | Tail logs (`backend`, `web-admin`, `ollama`) |
| `ragpack update` | Pull latest images and restart |
| `ragpack eject` | Copy `docker-compose.yml` locally for customization |

## Stack

| Layer | Technology |
|---|---|
| API server | Go + Fiber |
| Vector store | LanceDB (embedded) |
| Metadata / job state | SQLite |
| Embeddings | Ollama (local), OpenAI, or HuggingFace TEI |
| Admin UI | Next.js |

The stack is chosen for **low memory footprint** (~20MB idle), a **single static binary** with no runtime dependencies, and **fast query performance** — so RagPack runs comfortably on a $5 VPS, a Raspberry Pi, or a spare Mac Mini.

## Configuration

`.env.ragpack` (created by `ragpack init`):

```env
# Embedding provider
EMBED_PROVIDER=ollama          # ollama | openai | tei

# Ollama
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_EMBED_MODEL=nomic-embed-text

# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-3-small

# HuggingFace TEI
TEI_EMBED_MODEL=BAAI/bge-small-en-v1.5

# Ingestion
WORKER_COUNT=5
EMBED_RATE_LIMIT=10            # max embed API calls/sec
```

Storage defaults to `/data` inside the container, backed by a named Docker volume.

## API

Base URL: `http://localhost:9000/api/v1`

### Collections

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections` | List collections |
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:slug` | Get a collection |
| `DELETE` | `/collections/:slug` | Delete a collection and all its data |

```bash
curl -X POST http://localhost:9000/api/v1/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "My Docs"}'
```

### Ingest

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/ingest` | Ingest a URL or file upload |

```bash
# Ingest a URL
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -H "Content-Type: application/json" \
  -d '{"file_uri": "https://example.com/docs/guide"}'

# Upload a file
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -F "file=@./document.pdf"
```

Supported sources: `https://`, `s3://`, local file paths, file uploads.

### Query

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/query` | Semantic search |

```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Content-Type: application/json" \
  -d '{"query": "how do I configure authentication?", "top_k": 5}'
```

Response includes matched chunks with `chunk_text`, `file_uri`, `distance`, and `similarity` score (0–100).

### Documents

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections/:slug/documents` | List ingested documents (paginated) |
| `GET` | `/collections/:slug/documents/:id` | Get a document |
| `DELETE` | `/collections/:slug/documents/:id` | Delete a document and its chunks |
| `GET` | `/collections/:slug/documents/:id/chunks` | List all chunks |

## Admin UI

The admin UI at [http://localhost:3000](http://localhost:3000) lets you:

- Create and delete collections
- Ingest documents via URL or file upload
- Monitor ingestion status (ingesting / complete / failed)
- Delete individual documents
- Run queries against a collection

## Embedding models

Supported out of the box:

| Model | Dimensions | Provider |
|---|---|---|
| `nomic-embed-text` | 768 | Ollama |
| `mxbai-embed-large` | 1024 | Ollama |
| `bge-m3` | 1024 | Ollama |
| `all-minilm` | 384 | Ollama |
| `text-embedding-3-small` | 1536 | OpenAI |
| `text-embedding-3-large` | 3072 | OpenAI |
| `text-embedding-ada-002` | 1536 | OpenAI |

To add a model, edit `backend/pkg/embedder/dimensions.go`.

## Chunking

- **Markdown** — splits on headers, prepending parent breadcrumbs to each chunk for context
- **HTML** — converted to Markdown (stripping nav, footer, scripts) then chunked as above
- **PDF** — text extracted page-by-page, chunked as plain text
- **Plain text** — sliding window (2000 chars, 200 char overlap)

## Running locally

Prerequisites: Go 1.22+, Node 18+, Docker

**Backend** (hot reload via Air):

```bash
cd backend
./dev.sh
```

**Admin UI** (hot reload via Next.js):

```bash
cd web-admin
npm install
npm run dev
```

The backend dev server runs on `:9000`. The admin UI runs on `:3000` and proxies API requests to the backend.

For the embedding provider, the easiest local setup is Ollama — start it separately and point `OLLAMA_BASE_URL` at it in your `.env.ragpack`.

## License

MIT
