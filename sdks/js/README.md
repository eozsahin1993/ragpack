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

// Wait for ingestion to finish
const job = await collection.ingest(file);
await collection.jobs.waitUntilComplete(job.id);

// RAG — retrieve relevant chunks and get an LLM answer
const { answer, chunks } = await collection.rag({ query: "how does auth work?" });
console.log(answer);

// Semantic search (without LLM)
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
| `rag(options)`                      | Full RAG pipeline — retrieves chunks and returns an LLM answer |
| `findSimilar({ query, topK? })`     | Semantic search without LLM, returns ranked chunks |
| `jobs.list()`                       | List ingestion jobs for this collection  |
| `jobs.get(id)`                      | Get a single job by ID                   |
| `jobs.waitUntilComplete(id)`        | Poll until job is `complete` or `failed` |
| `documents.list(options?)`          | List indexed documents                   |
| `documents.delete(id)`              | Delete a document and its chunks         |

#### `rag(options)`

Runs the full RAG pipeline server-side: embeds the query, retrieves the top-K chunks, fills the prompt template, and calls the configured LLM.

```ts
const { answer, chunks } = await collection.rag({
  query: "How do I reset my password?",
  // All options below are optional
  promptSlug: "basic_rag",   // defaults to "basic_rag" if omitted
  topK: 5,                   // number of chunks to retrieve (default: 5)
  model: "gpt-4o",           // LLM model; falls back to server default
  minSimilarity: 70,         // drop chunks below this similarity score (0–100)
});

console.log(answer);

for (const chunk of chunks) {
  console.log(chunk.similarity, chunk.chunk_text);
}
```

#### Prompt templates

Three built-in prompts are always available: `basic_rag`, `rag_with_citations`, and `concise_rag`. List all available prompts (including any you've created in the admin UI) with:

```ts
const prompts = await client.prompts.list();
// [{ slug: "basic_rag", name: "Basic RAG", is_system: true, ... }, ...]
```

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
