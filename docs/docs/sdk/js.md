---
sidebar_position: 1
---

# JavaScript / TypeScript SDK

The `ragpack-js` package is a typed client for the RagPack REST API.

## Install

```bash
npm install ragpack-js
```

## Setup

```ts
import { RagPack } from "ragpack-js";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
  apiKey: "rp_...",
});
```

## Collections

```ts
const col = await client.collections.create("my-docs");
const collections = await client.collections.list();
await client.collections.delete("my-docs");
```

## Ingest

```ts
const collection = client.collection("my-docs");

// Upload a file
await collection.ingest(file);

// Ingest from a URL
await collection.ingest({ uri: "https://example.com/docs/guide" });

// Ingest from S3
await collection.ingest({ uri: "s3://bucket/report.pdf", mimeType: "application/pdf" });
```

## Semantic search

```ts
const results = await collection.findSimilar({
  query: "how do I configure authentication?",
  topK: 5,
});

for (const r of results) {
  console.log(r.similarity, r.chunkText);
}
```

## RAG (retrieve + generate)

```ts
const { answer, chunks } = await collection.rag({
  query: "How do I reset my password?",
  topK: 5,
  model: "gpt-4o",        // falls back to server default if omitted
  minSimilarity: 60,      // filter out low-relevance chunks
});
```

## Jobs & documents

```ts
// Check ingestion status
const jobs = await collection.jobs.list();

// List indexed documents
const docs = await collection.documents.list();

// Delete a document and its chunks
await collection.documents.delete(documentId);
```

## Error handling

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
