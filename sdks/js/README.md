# RagPack

TypeScript SDK for [RagPack](https://github.com/eozsahin1993/ragpack) — a self-hosted semantic search and RAG engine.

## Installation

```bash
npm install ragpack-js
```

## Quick start

```ts
import { RagPack } from "ragpack-js";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
  apiKey: "rp_...",
});

// Create a collection
const collection = await client.collections.create("my-docs");

// Ingest a file
await collection.ingest(file);

// Or ingest from a URL
await collection.ingest({ uri: "https://example.com/docs/guide" });

// Or from S3
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
  console.log(r.vector_similarity, r.chunk_text);
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
| `ingest({ uri, mimeType? })`        | Ingest from a URL or S3 URI, MIME type auto-detected if omitted |
| `rag(options)`                      | Full RAG pipeline — retrieves chunks and returns an LLM answer |
| `findSimilar(options)`              | Hybrid (vector + keyword) search without LLM, returns ranked chunks |
| `jobs.list()`                       | List ingestion jobs for this collection  |
| `jobs.get(id)`                      | Get a single job by ID                   |
| `jobs.waitUntilComplete(id)`        | Poll until job is `complete` or `failed` |
| `documents.list(options?)`          | List indexed documents                   |
| `documents.get(id)`                 | Get a single document by ID              |
| `documents.metadata(id)`            | Get a document's typed metadata field values |
| `documents.rename(id, name)`        | Set the display name of a document       |
| `documents.update(id, options)`     | Update a document's name, `extraJson`, and/or typed metadata |
| `documents.delete(id)`              | Delete a document and its chunks         |

#### `rag(options)`

Runs the full RAG pipeline server-side: embeds the query, retrieves the top-K chunks, fills the prompt template, and calls the configured LLM.

```ts
const { answer, chunks } = await collection.rag({
  query: "How do I reset my password?",
  // All options below are optional
  promptSlug: "basic_rag",   // defaults to "basic_rag" if omitted
  topK: 2,                   // number of chunks to retrieve (default: 2)
  model: "gpt-4o",           // LLM model; falls back to server default
  minSimilarity: 70,         // drop chunks below this similarity score (0–100)
  filters: { mime_type: "application/pdf" },
});

console.log(answer);

for (const chunk of chunks) {
  console.log(chunk.vector_similarity, chunk.chunk_text);
}
```

#### Hybrid (vector + keyword) search

`findSimilar` and `rag` run hybrid search by default: a vector pass and a keyword/BM25 pass are fused server-side via weighted RRF (semantic-favored 7:3). Each result carries `vector_similarity`, `keyword_bm25_score`, `rrf_score`, and `rrf_score_normalized` (0–100, normalized so the batch's top result is always 100).

```ts
const results = await collection.findSimilar({
  query: "password reset",
  vectorSearchOnly: false,          // set true to skip the keyword pass entirely
  hybridSettings: {
    semanticWeight: 0.7,
    fullTextWeight: 0.3,
    rrfK: 60,
    fullTextLimit: 200,
  },
});
```

#### Filtering by metadata

Both `findSimilar` and `rag` accept a `filters` option — a MongoDB-style expression compiled server-side into a predicate over a document's built-in (`mime_type`, `source_name`, `file_uri`, `created_at`, `updated_at`, …) and registered metadata fields.

```ts
await collection.findSimilar({
  query: "refund policy",
  filters: {
    $and: [
      { mime_type: "application/pdf" },
      { $or: [{ tags: { $contains: "billing" } }, { created_at: { $gte: "7 days ago" } }] },
    ],
  },
});
```

Supported operators: `$eq`, `$ne`, `$gt`, `$gte`, `$lt`, `$lte`, `$in`, `$nin`, `$exists`, `$like`/`$ilike` (string fields), `$contains`/`$containsAny`/`$containsAll` (array fields), plus `$and`/`$or` for nesting.

#### Prompt templates

Three built-in prompts are always available: `basic_rag`, `rag_with_citations`, and `concise_rag`. List all available prompts (including any you've created in the admin UI) with:

```ts
const prompts = await client.prompts.list();
// [{ slug: "basic_rag", name: "Basic RAG", is_system: true, ... }, ...]
```

## Error handling

All methods throw `RagPackError` on non-2xx responses:

```ts
import { RagPack, RagPackError } from "ragpack-js";

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

## License

MIT
