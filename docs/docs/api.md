---
sidebar_position: 2
---

# API Reference

Base URL: `http://localhost:9000/api/v1`

## Collections

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections` | List all collections |
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:slug` | Get a collection |
| `DELETE` | `/collections/:slug` | Delete a collection and all its data |

```bash
curl -X POST http://localhost:9000/api/v1/collections \
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
  -H "Content-Type: application/json" \
  -d '{"file_uri": "https://example.com/docs/guide"}'

# Upload a file
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -F "file=@./document.pdf"
```

Supported sources: `https://`, `s3://`, local file paths, file uploads.  
Supported formats: `.txt`, `.md`, `.html`, `.pdf`

## Query

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/query` | Semantic search |

```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Content-Type: application/json" \
  -d '{"query": "how do I configure authentication?", "top_k": 5}'
```

Response includes matched chunks with `chunk_text`, `file_uri`, `distance`, and `similarity` (0–100).

## Documents

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections/:slug/documents` | List ingested documents (paginated) |
| `GET` | `/collections/:slug/documents/:id` | Get a document |
| `DELETE` | `/collections/:slug/documents/:id` | Delete a document and its chunks |
| `GET` | `/collections/:slug/documents/:id/chunks` | List all chunks |
