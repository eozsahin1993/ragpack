<p align="center">
  <img src="web-admin/public/logo-with-text.svg" width="320" alt="RagPack" />
</p>

<p align="center">
  Open-source, self-hosted RAG infrastructure built for early-stage startups. High performance, low cost, and simple to run.
</p>

## What it does

- Ingest documents from URLs, file uploads, or S3. RagPack handles fetching, parsing, chunking, and embedding automatically
- **Semantic search**: query a collection with natural language and get back ranked, scored chunks ready to use in any LLM prompt
- **RAG**: send a question with a prompt template and a model. RagPack retrieves the relevant chunks and returns a grounded answer
- Bring your own embedding model. Ollama or TEI for fully local inference, or any OpenAI-compatible provider
- Manage everything via REST API or the built-in admin UI

Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`, `.csv`, `.json`, `.xml`

## Who it's for

Developers who want to add semantic search or RAG to an existing app without building and maintaining the infrastructure themselves. Designed to run on minimal infrastructure. Spin it up with `npx ragpack start`, then use the TypeScript SDK or REST API. No pipeline code, no boilerplate, no cloud bill.

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

With OpenAI, edit `.env.ragpack` first:

```env
EMBED_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-3-small
```

Then `npx ragpack start`.

The admin UI opens automatically at [http://localhost:3000](http://localhost:3000). Logs stream in the terminal. Ctrl+C stops following logs without stopping the stack.

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

The stack is chosen for **low memory footprint** (~20MB idle), a **single static binary** with no runtime dependencies, and **fast query performance**. RagPack runs comfortably on a $5 VPS or a spare Mac Mini.

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

## TypeScript SDK

```bash
npm install ragpack-js
```

```ts
import { RagPack } from "ragpack-js";

const client = new RagPack({ baseUrl: "http://localhost:9000", apiKey: "rp_..." });
const collection = client.collection("my-docs");

// Ingest a file
await collection.ingest(file);

// Semantic search: returns ranked chunks (hybrid by default, with optional filters)
const results = await collection.findSimilar({
  query: "what is RagPack?",
  filters: { category: "docs" },
});

// RAG: retrieves chunks and returns an LLM answer
const { answer, chunks } = await collection.rag({
  query: "How do I configure authentication?",
  promptSlug: "basic-rag",
  model: "gpt-4o",
});

// Documents: fetch, read typed metadata, or update name/extra_json/metadata
const doc = await collection.documents.get(docId);
const metadata = await collection.documents.metadata(docId);
await collection.documents.update(docId, { metadata: { status: "published" } });
```

For the full SDK reference, see [ragpack.dev/docs/sdk/js](https://ragpack.dev/docs/sdk/js).

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
Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`, `.csv`, `.json`, `.xml`

### Query

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/query` | Semantic or hybrid search, returns ranked chunks |
| `POST` | `/collections/:slug/rag` | RAG, retrieves chunks and returns an LLM answer |

```bash
# Semantic search (hybrid by default — vector + keyword, merged with weighted RRF)
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query": "how do I configure authentication?", "top_k": 5, "filters": {"category": "docs"}}'

# RAG
curl -X POST http://localhost:9000/api/v1/collections/my-docs/rag \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query": "how do I configure authentication?", "top_k": 5, "prompt_slug": "my-prompt", "model": "llama3"}'
```

The query response includes `chunk_text`, `file_uri`, `vector_distance`, and `vector_similarity` (0–100), plus `keyword_bm25_score`, `rrf_score`, and `rrf_score_normalized` when hybrid search ran. Pass `vector_search_only: true` to skip the keyword pass. The RAG response adds `answer` and `formatted_prompt`.

### Documents

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections/:slug/documents` | List ingested documents (paginated) |
| `GET` | `/collections/:slug/documents/:id` | Get a document |
| `GET` | `/collections/:slug/documents/:id/metadata` | Get a document's typed metadata field values |
| `PATCH` | `/collections/:slug/documents/:id` | Update `name`, `extra_json`, and/or `metadata` |
| `DELETE` | `/collections/:slug/documents/:id` | Delete a document and its chunks |
| `GET` | `/collections/:slug/documents/:id/chunks` | List all chunks |

### API keys

| Method | Path | Description |
|---|---|---|
| `GET` | `/keys` | List API keys |
| `POST` | `/keys` | Create an API key |
| `DELETE` | `/keys/:id` | Delete an API key |

Keys carry two independent, fail-closed grant kinds: per-collection `grants` (`read`/`write`/`both`, or wildcard by omitting `collection_slug`) and `admin_grants` for instance administration (`keys`/`prompts`/`collections`/`*`) — so a key that can manage prompts doesn't automatically get collection access, and vice versa. See the [API reference](https://ragpack.dev/docs/api) for the full grant model and examples.

## Admin UI

The admin UI at [http://localhost:3000](http://localhost:3000) lets you:

- Create and delete collections
- Ingest documents via URL or file upload
- Monitor ingestion status (ingesting / complete / failed)
- Delete individual documents
- Run queries against a collection, with hybrid search and filters
- Manage API keys and their collection/admin grants

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

- **Markdown**: splits on headers, prepending parent breadcrumbs to each chunk for context
- **HTML**: converted to Markdown (stripping nav, footer, scripts) then chunked as above
- **PDF**: text extracted page-by-page, chunked as plain text
- **Plain text**: sliding window (2000 chars, 200 char overlap)

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
