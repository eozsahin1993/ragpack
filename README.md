# RagPack

Self-hostable semantic search and RAG infrastructure. High-performance, low-cost, and up in minutes. Bring your own AI — Ollama for local models or any OpenAI-compatible provider.

## What it does

- Ingest documents from URLs, file uploads, or S3 — RagPack fetches, parses, chunks, and embeds them automatically
- Organize documents into **collections** and query them with natural language
- Get back ranked chunks with similarity scores, ready to drop into any LLM prompt
- Manage everything via REST API or the built-in admin UI

Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`, `.csv`, `.json`

## Quick start

**Prerequisites:** [Docker](https://docs.docker.com/get-docker/) must be installed and running.

Install the CLI:

```bash
npm install -g ragpack
```

Or run without installing via `npx`:

```bash
npx ragpack init
```

Then start the stack:

```bash
ragpack init       # creates .env.ragpack in the current directory
ragpack start      # starts the stack (API on :9000, admin UI on :3000)
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

The admin UI opens automatically at [http://localhost:3000](http://localhost:3000). Logs stream in the terminal — Ctrl+C stops following logs without stopping the stack.

For headless/VPS use, start in the background:

```bash
ragpack start --detach
```

## CLI

| Command | Description |
|---|---|
| `ragpack init` | Create `.env.ragpack` in the current directory |
| `ragpack start [--profile ollama]` | Start the stack (streams logs, opens admin UI) |
| `ragpack start --detach` | Start in background without following logs |
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

All requests require an API key. On first startup the backend prints it to the logs and saves it to `/data/api_key` inside the container. To retrieve it:

```bash
ragpack logs backend | grep "Key:"
```

Pass it as a bearer token in every request:

```bash
export RAGPACK_API_KEY=rp_...
```

### Collections

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections` | List collections |
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:slug` | Get a collection |
| `DELETE` | `/collections/:slug` | Delete a collection and all its data |

```bash
curl -X POST http://localhost:9000/api/v1/collections \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
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
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"file_uri": "https://example.com/docs/guide"}'

# Upload a file
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -F "file=@./document.pdf"
```

Supported sources: `https://`, `s3://`, file uploads.
Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`, `.csv`, `.json`, `.docx`, `.pptx`, `.xlsx`

### Query

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/query` | Semantic search |

```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
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

## Contributing / local development

Prerequisites: Docker

```bash
npm run dev           # backend + admin UI with hot reload
npm run dev:ollama    # include Ollama
npm run dev:tei       # include HuggingFace TEI
```

Backend runs on `:9000`, admin UI on `:3000`.

## License

Apache-2.0
