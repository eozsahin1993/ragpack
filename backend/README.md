# RagPack Backend

Go + Fiber HTTP server for the RagPack engine. Uses LanceDB (via CGO) for vector storage and SQLite for metadata.

## Running with Docker (recommended)

Requires [Docker Desktop](https://www.docker.com/products/docker-desktop/) running locally.

**1. Copy and fill in your env file:**

```bash
cp .env.example .env
```

**2a. Using OpenAI embeddings:**

Set `OPENAI_API_KEY` and `OPENAI_EMBED_MODEL` in `.env`, then:

```bash
docker compose up
```

**2b. Using Ollama (local, no API key):**

```bash
docker compose --profile ollama up

# Pull the embedding model once (first time only):
docker compose exec ollama ollama pull nomic-embed-text
```

Server starts on `http://localhost:9000`.

Data (SQLite + LanceDB) is persisted in a Docker volume — it survives restarts.

---

## Running locally (macOS arm64 only)

```bash
cd backend
./dev.sh
```

Downloads the LanceDB native library on first run, then starts the server.

---

## API

All routes are prefixed with `/api/v1`.

### Health

```
GET /api/v1/health
```

### Collections

```
POST   /api/v1/collections              Create a collection
GET    /api/v1/collections              List all collections
GET    /api/v1/collections/:slug        Get by slug
GET    /api/v1/collections/id/:id       Get by ID
PATCH  /api/v1/collections/id/:id       Update name
DELETE /api/v1/collections/:slug        Delete by slug
DELETE /api/v1/collections/id/:id       Delete by ID
```

**Create body:**
```json
{
  "name": "company_wiki",
  "embed_model": "text-embedding-3-small",
  "vector_dim": 1536
}
```

### Jobs

```
GET /api/v1/jobs                                        All jobs (global)
GET /api/v1/jobs?status=pending                         Global filter by status
GET /api/v1/jobs/:id                                    Get a job by ID
GET /api/v1/collections/:slug/jobs                      Jobs for a collection
GET /api/v1/collections/:slug/jobs?status=pending       Filter by status
```

Valid statuses: `pending`, `processing`, `complete`, `failed`

### Ingest

```
POST /api/v1/collections/:slug/ingest
```

**File upload (multipart):**
```bash
curl -X POST http://localhost:9000/api/v1/collections/my-wiki/ingest \
  -F "file=@document.txt"
```

**URI-based (S3, HTTP, local):**
```json
{ "file_uri": "s3://my-bucket/doc.pdf", "mime_type": "application/pdf" }
```

Returns `202 Accepted` with the job object. Processing happens asynchronously.

### Query

```
POST /api/v1/collections/:slug/query
```

```json
{ "query": "what is the refund policy?", "top_k": 10 }
```
