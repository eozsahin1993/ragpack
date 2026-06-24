# ragpack

Self-hostable semantic search and RAG (Retrieval-Augmented Generation) infrastructure for developers. Bring your own AI — Ollama for local models or OpenAI — and be up and running in minutes.

## What it is

ragpack gives you a REST API for ingesting documents and running semantic search over them. You organize documents into **collections**, ingest files or URLs, and query with natural language. It handles chunking, embedding, and vector storage so you don't have to.

Use it as the retrieval layer in a RAG pipeline, as a semantic search backend, or as a knowledge base for an AI assistant.

## How it works

```
Document (URL / file upload / S3)
        ↓
   Fetch & parse
        ↓
   Chunk (Markdown-aware / HTML / PDF / plain text)
        ↓
   Embed (Ollama or OpenAI)
        ↓
   Store in LanceDB (vectors) + SQLite (metadata)
        ↓
   Query via REST API
```

State is tracked in SQLite. Vectors live in LanceDB on disk. No external services required beyond your embedding provider.

## Stack

| Layer | Technology |
|---|---|
| API server | Go + Fiber |
| Vector store | LanceDB (embedded) |
| Metadata / job state | SQLite |
| Embeddings | Ollama (local) or OpenAI |
| Admin UI | Next.js |

### Why Go?

ragpack is designed to run on a single small server — a $5 VPS, a Raspberry Pi, a spare Mac Mini. Go fits that constraint well:

- **Low memory footprint** — the server idles at ~20MB RAM, leaving headroom for the embedding model.
- **Single static binary** — no Python virtualenv, no JVM, no runtime to manage. The Docker image is small and the binary starts in milliseconds.
- **Built-in concurrency** — the ingestion worker pool uses goroutines and channels. Embedding batches run concurrently without the complexity of async frameworks.
- **Simple builds** — pure Go where possible means no CGo toolchain requirements and reproducible Docker builds.

### Why LanceDB?

Most vector databases are external services — Pinecone is SaaS, Weaviate and Qdrant need their own container with memory and a port. For a self-hosted tool that should be "up in minutes", that's friction.

LanceDB is an **embedded** vector store — it runs inside the ragpack process, reads and writes directly to disk, and needs no separate process, network port, or ops attention. Characteristics that matter here:

- **Zero ops** — nothing to provision, monitor, or restart separately.
- **Disk-persistent** — survives container restarts without re-indexing.
- **Columnar storage (Apache Arrow)** — fast filtered scans, efficient memory use.
- **Rust core** — competitive query performance without running a separate service.
- **Scales for single-node** — handles millions of vectors on a modest machine, which covers the vast majority of self-hosted RAG use cases.

## Quick start

**With Ollama (fully local):**

```bash
cp .env.example .env
docker compose --profile ollama up -d
```

Then open [http://localhost:3000](http://localhost:3000) to access the admin UI.

**With OpenAI:**

```bash
cp .env.example .env
# Set EMBED_PROVIDER=openai, OPENAI_API_KEY, OPENAI_EMBED_MODEL in .env
docker compose up -d
```

## Configuration

Copy `.env.example` to `.env` and edit:

```env
# Choose your embedding provider
EMBED_PROVIDER=ollama          # or: openai

# Ollama (used when EMBED_PROVIDER=ollama)
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_EMBED_MODEL=nomic-embed-text

# OpenAI (used when EMBED_PROVIDER=openai)
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-3-small

# Ingestion workers
WORKER_COUNT=5
EMBED_RATE_LIMIT=10            # max embed API calls/sec
```

Storage paths default to `/data` inside the container, backed by a named Docker volume.

## API

Base URL: `http://localhost:9000/api/v1`

### Collections

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections` | List collections |
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:slug` | Get a collection |
| `DELETE` | `/collections/:slug` | Delete a collection and all its data |

**Create a collection** — embed model is auto-selected from the configured provider:
```bash
curl -X POST http://localhost:9000/api/v1/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "My Docs"}'
```

### Ingest

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/ingest` | Ingest a URL or file upload |

**Ingest a URL** — MIME type is auto-detected:
```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -H "Content-Type: application/json" \
  -d '{"file_uri": "https://example.com/docs/guide"}'
```

**Upload a file:**
```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -F "file=@./document.pdf"
```

Supported sources: `https://`, `s3://`, local file paths, file uploads.  
Supported formats: `.txt`, `.md`, `.html`, `.pdf`.

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
| `GET` | `/collections/:slug/documents/:id/chunks` | List all chunks (debug) |

## Admin UI

The admin UI runs at [http://localhost:3000](http://localhost:3000) and lets you:

- Create and delete collections
- Ingest documents via URL or file upload
- View ingestion status (ingesting / complete / failed) with error details
- Delete individual documents
- Run queries against a collection

## Embedding models

Vector dimensions are looked up automatically from the model name — you don't need to specify them. Supported models include:

| Model | Dimensions | Provider |
|---|---|---|
| `nomic-embed-text` | 768 | Ollama |
| `mxbai-embed-large` | 1024 | Ollama |
| `bge-m3` | 1024 | Ollama |
| `all-minilm` | 384 | Ollama |
| `text-embedding-3-small` | 1536 | OpenAI |
| `text-embedding-3-large` | 3072 | OpenAI |
| `text-embedding-ada-002` | 1536 | OpenAI |

To add a model not in this list, add it to `backend/pkg/embedder/dimensions.go`.

## Chunking

Documents are split before embedding:

- **Markdown** — splits on headers (`#`, `##`, `###`…) first, with parent header breadcrumbs prepended to each chunk for context. Falls back to paragraph then character splitting for oversized sections.
- **HTML** — converted to Markdown (stripping nav, footer, scripts, ads) then chunked as above.
- **PDF** — text extracted page-by-page, then chunked as plain text.
- **Plain text** — sliding window (2000 chars, 200 char overlap).

Default chunk size and overlap can be configured in `backend/pkg/chunker/chunker.go`.

## License

MIT
