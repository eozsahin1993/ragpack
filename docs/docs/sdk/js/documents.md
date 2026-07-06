---
sidebar_position: 7
---

# Documents

Once a job completes, the ingested file becomes a **document** — a record of everything RagPack indexed from that file. You can list, rename, and delete documents via `collection.documents`.

## List documents

```ts
const docs = await collection.documents.list();

for (const doc of docs) {
  console.log(doc.name ?? doc.fileUri, doc.chunkCount, doc.status);
}
```

Results are paginated. Use `limit` and `offset` to page through large collections:

```ts
const page2 = await collection.documents.list({ limit: 20, offset: 20 });
```

| Option | Type | Description |
|---|---|---|
| `limit` | `number` | Max documents to return. Defaults to `50`. |
| `offset` | `number` | Pagination offset. Defaults to `0`. |

Each document contains:

| Field | Type | Description |
|---|---|---|
| `id` | `string` | Unique document ID |
| `name` | `string \| undefined` | Display name. Set at ingest time or via `rename()`. |
| `fileUri` | `string` | Source URI of the ingested file |
| `mimeType` | `string` | Detected MIME type |
| `chunkCount` | `number` | Number of chunks indexed |
| `status` | `string` | `ingesting` \| `complete` \| `failed` |
| `error` | `string \| undefined` | Error detail if `status` is `failed` |
| `createdAt` | `string` | ISO 8601 timestamp |

## Rename a document

Sets the display `name` on a document. Useful for giving a human-readable label to a file ingested from a URI.

```ts
await collection.documents.rename(doc.id, "Q4 2024 Report");
```

## Delete a document

Removes the document and all its chunks from the index. This is irreversible.

```ts
await collection.documents.delete(doc.id);
```
