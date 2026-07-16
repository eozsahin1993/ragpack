<p align="center">
  <img src="web-admin/public/logo-with-text.svg" width="320" alt="RagPack" />
</p>

<p align="center">
  Open-source, self-hosted RAG infrastructure built for performance and low cost.
</p>

<p align="center">
  <a href="https://www.npmjs.com/package/ragpack"><img src="https://img.shields.io/npm/v/ragpack.svg" alt="npm version" /></a>
  <a href="https://www.npmjs.com/package/ragpack"><img src="https://img.shields.io/npm/dm/ragpack.svg" alt="npm downloads" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/eozsahin1993/ragpack.svg" alt="License" /></a>
</p>

## Why RagPack

Want to add RAG to your app? Most stacks reach for LangChain and Pinecone, fast to get a demo running, but the catch is what they cost to run and maintain as you grow. A RAG solution shouldn't cost a fortune.

RagPack is a single Go binary instead. Storage is embedded directly in the process, LanceDB for vectors and SQLite for metadata, so there's nothing else to run, and Go's efficiency at I/O and concurrent processing means it's light enough to run comfortably on something as small as an EC2 t3.micro instance instead of a costly managed vector DB.

## Features

- Semantic search and RAG endpoints with prompts baked in
- Hybrid retrieval: BM25 keyword search + vector search, merged with Reciprocal Rank Fusion (RRF) using customizable weights
- Bring your own embedding model, Ollama or TEI for fully local inference, or any OpenAI-compatible provider
- Ingest documents from URLs, file uploads, or S3
- Client SDKs (JS/TS) for dropping into an existing app
- Mongo-style filters on custom document properties
  - i.e: `{"$and": [{"category": {"$in": ["research", "legal"]}}, {"score": {"$gt": 0.8}}]}`
- Chunking strategy picked per file type:
  - Context aware: preserves headers as breadcrumbs
  - Paragraph: splits on paragraph boundaries
  - Sliding window: fixed-size windows with overlap (2000 chars, 200 char overlap)
  - Row group: keeps row headers attached to every row, for CSV/XLS
  - Auto: picks the right strategy from the file's mime type
- Smart refresh on a timer, only re-embeds and re-inserts chunks that actually changed
- Admin dashboard for managing collections, documents, and queries
- Built-in analytics for embedding/LLM costs, usage metrics, and query evaluations

Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`, `.csv`, `.json`, `.xml`

## RAGAS evaluations

Scored against [SQuAD 2.0](https://rajpurkar.github.io/SQuAD-explorer/) (30 questions, real questions and answers, not RagPack's own docs) through the actual `/rag` endpoint, judged by `gpt-4o-mini`:

| Metric | Score |
|---|---|
| Faithfulness | 0.94 |
| Answer relevancy | 0.81 |
| Context precision | 0.97 |
| Context recall | 1.00 |

Retrieval quality is also checked separately against [BEIR](https://github.com/beir-cellar/beir)'s SciFact benchmark: `nDCG@10` 0.87, `Recall@100` 1.00.

Reproducible with the eval harness in [`eval/`](eval/): `python3 eval/run_eval.py --api-key <key> --model gpt-4o-mini`.

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
