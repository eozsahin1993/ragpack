---
sidebar_position: 3
---

# Ingesting Documents

Ingestion is the process of taking a document, parsing it into chunks, embedding each chunk, and storing the vectors for search. You kick it off with a single call — RagPack handles the rest asynchronously.

Ingest returns a `Job` immediately. The job runs in the background; use [Jobs](./jobs.md) to track its status.

## Upload a file

Pass a `File` or `Blob` directly. The MIME type is detected from the filename automatically.

```ts
const collection = client.collection("my-docs");

// From a file input (browser)
const file = input.files[0];
const job = await collection.ingest(file);

// From a Buffer (Node.js)
import { readFileSync } from "fs";
const buf = readFileSync("report.pdf");
const job = await collection.ingest(new Blob([buf]), "report.pdf");
```

## Ingest from a URL

```ts
const job = await collection.ingest({
  uri: "https://example.com/docs/guide",
});
```

## Ingest from S3

```ts
const job = await collection.ingest({
  uri: "s3://my-bucket/reports/q4.pdf",
  mimeType: "application/pdf",
});
```

AWS credentials are picked up from the environment (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`) or the default credential chain on the server.

## Supported formats

`.txt`, `.md`, `.html`, `.pdf`, `.docx`, `.pptx`, `.xlsx`, `.csv`, `.json`, `.xml`

## Rate limiting and concurrency

Ingestion throughput is controlled server-side via `.env.ragpack`. You don't need to throttle calls from the SDK — the server queues jobs and processes them with a fixed worker pool.

| Variable | Default | Description |
|---|---|---|
| `WORKER_COUNT` | `5` | Number of documents processed concurrently |
| `EMBED_RATE_LIMIT` | `10` | Max embedding API calls per second |

`EMBED_RATE_LIMIT` is particularly important when using an external provider like OpenAI. Set it to stay within your API tier's rate limit and avoid `429` errors during large ingestion runs.

## Waiting for ingestion to complete

By default `ingest` returns as soon as the job is queued. If you need to wait before querying, use `waitUntilComplete`:

```ts
const job = await collection.ingest(file);
await collection.jobs.waitUntilComplete(job.id);

// Safe to query now
const results = await collection.findSimilar({ query: "..." });
```

`waitUntilComplete` throws a `RagPackError` with status `422` if the job fails, so you can catch it:

```ts
import { RagPackError } from "ragpack-js";

try {
  const job = await collection.ingest(file);
  await collection.jobs.waitUntilComplete(job.id);
} catch (err) {
  if (err instanceof RagPackError) {
    console.error("ingestion failed:", err.message);
  }
}
```

## Handling failure without waiting

If you ingest without `waitUntilComplete`, failures are silent until you check the job. Poll the job status manually or inspect the job list:

```ts
const job = await collection.ingest(file);

// Check later
const updated = await collection.jobs.get(job.id);

if (updated.status === "failed") {
  console.error("ingestion failed:", updated.error);
}
```

Possible statuses:

| Status | Meaning |
|---|---|
| `pending` | Queued, not yet started |
| `processing` | Actively fetching, parsing, chunking, and embedding |
| `complete` | Indexed and ready to query |
| `failed` | Something went wrong — `job.error` contains the reason |
