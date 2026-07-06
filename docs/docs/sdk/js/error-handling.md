---
sidebar_position: 8
---

# Error Handling

All SDK methods throw a `RagPackError` when the server returns a non-2xx response or when a job fails.

## Catching errors

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

## RagPackError

| Property | Type | Description |
|---|---|---|
| `status` | `number` | HTTP status code returned by the server |
| `message` | `string` | Human-readable error message |

## Common error codes

| Status | Cause |
|---|---|
| `401` | Missing or invalid API key |
| `404` | Collection or document not found |
| `408` | `waitUntilComplete` timed out before the job finished |
| `422` | Ingestion job failed — check `job.error` for details |
| `500` | Unexpected server error |
