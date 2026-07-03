---
sidebar_position: 4
---

# API Reference

Base URL: `http://localhost:9000/api/v1`

All requests require an API key. On first startup the backend prints it to the logs and saves it to `/data/api_key` inside the container. To retrieve it:

```bash
ragpack logs backend | grep "Key:"
```

Pass it as a bearer token in every request:

```bash
export RAGPACK_API_KEY=rp_...
```

## Collections

| Method | Path | Description |
|---|---|---|
| `GET` | `/collections` | List all collections |
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:slug` | Get a collection |
| `DELETE` | `/collections/:slug` | Delete a collection and all its data |

```bash
curl -X POST http://localhost:9000/api/v1/collections \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
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
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"file_uri": "https://example.com/docs/guide"}'

# Upload a file
curl -X POST http://localhost:9000/api/v1/collections/my-docs/ingest \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
  -F "file=@./document.pdf"
```

Supported sources: `https://`, `s3://`, file uploads.  
Supported formats: `.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`

## Query

| Method | Path | Description |
|---|---|---|
| `POST` | `/collections/:slug/query` | Semantic search |

```bash
curl -X POST http://localhost:9000/api/v1/collections/my-docs/query \
  -H "Authorization: Bearer $RAGPACK_API_KEY" \
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
