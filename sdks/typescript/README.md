# RagPack

TypeScript SDK for [RagPack](https://github.com/eozsahin1993/ragpack) — a self-hosted semantic search and RAG engine.

## Installation

```bash
npm install ragpack
```

## Quick start

```ts
import { RagPack } from "ragpack";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
  apiKey: "rp_...",
});

// Create a collection
const collection = await client.collections.create("my-docs");

// Ingest a file
await collection.ingest(file);

// Or ingest from a remote URI
await collection.ingest({ uri: "s3://my-bucket/report.pdf" });

// Search
const results = await collection.findSimilar({ query: "how does auth work?", topK: 5 });

for (const r of results) {
  console.log(r.similarity, r.chunk_text);
}
```

## API

### `new RagPack({ baseUrl, apiKey })`

| Option    | Description                                          |
|-----------|------------------------------------------------------|
| `baseUrl` | URL of your RagPack backend (e.g. `http://localhost:9000`) |
| `apiKey`  | API key printed by the backend on first startup      |

### `client.collections`

| Method                              | Description                              |
|-------------------------------------|------------------------------------------|
| `list()`                            | List all collections                     |
| `create(name, options?)`            | Create a collection, returns a scoped client |
| `get(slug)`                         | Get collection metadata by slug          |
| `delete(slug)`                      | Delete a collection and all its documents |

### `client.collection(slug)`

Returns a `CollectionClient` scoped to that collection.

| Method                              | Description                              |
|-------------------------------------|------------------------------------------|
| `ingest(file, filename?)`           | Upload a file directly                   |
| `ingest({ uri, mimeType? })`        | Ingest from a remote URI, MIME type auto-detected if omitted |
| `findSimilar({ query, topK? })`     | Semantic search, returns ranked results  |
| `jobs.list()`                       | List ingestion jobs                      |
| `jobs.get(id)`                      | Get a single job by ID                   |
| `documents.list(options?)`          | List indexed documents                   |
| `documents.delete(id)`              | Delete a document and its chunks         |

## Error handling

All methods throw `RagPackError` on non-2xx responses:

```ts
import { RagPack, RagPackError } from "ragpack";

try {
  await collection.findSimilar({ query: "..." });
} catch (err) {
  if (err instanceof RagPackError) {
    console.error(err.status, err.message);
  }
}
```

## Requirements

- Node.js 18+
- A running RagPack backend ([setup guide](https://github.com/eozsahin1993/ragpack))
