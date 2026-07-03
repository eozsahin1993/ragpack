---
sidebar_position: 1
---

# Setup

The `ragpack-js` package is the official TypeScript client for RagPack. It wraps the REST API with a typed, ergonomic interface — no need to write `fetch` calls manually.

## Install

```bash
npm install ragpack-js
```

## Initialize the client

```ts
import { RagPack } from "ragpack-js";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
  apiKey: "rp_...",
});
```

| Option | Type | Description |
|---|---|---|
| `baseUrl` | `string` | Base URL of your RagPack backend |
| `apiKey` | `string` | API key printed to the logs on first startup |

To retrieve your API key:

```bash
ragpack logs backend | grep "Key:"
```

## Scoping to a collection

Most operations are performed on a specific collection. Use `client.collection(slug)` to get a scoped client rather than passing the slug to every call:

```ts
const collection = client.collection("my-docs");

await collection.ingest(file);
const results = await collection.findSimilar({ query: "..." });
```
