---
sidebar_position: 7
---

# Documents

Once a job completes, the ingested file becomes a **document** — a record of everything RagPack indexed from that file. You can list, get, rename, update, and delete documents via `collection.documents`.

## List documents

```ts
const docs = await collection.documents.list();

for (const doc of docs) {
  console.log(doc.name ?? doc.file_uri, doc.chunk_count, doc.status);
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
| `name` | `string \| undefined` | Display name. Set at ingest time or via `rename()`/`update()`. |
| `file_uri` | `string` | Source URI of the ingested file |
| `mime_type` | `string` | Detected MIME type |
| `external_id` | `string \| undefined` | Caller-supplied ID for correlating with an external system, set at ingest time |
| `extra_json` | `string \| undefined` | Freeform JSON metadata, set at ingest time or via `update()` |
| `chunk_count` | `number` | Number of chunks indexed |
| `status` | `string` | `ingesting` \| `complete` \| `failed` |
| `error` | `string \| undefined` | Error detail if `status` is `failed` |
| `created_at` | `string` | ISO 8601 timestamp |

Response fields are returned exactly as the API sends them (snake_case).

## Get a document

```ts
const doc = await collection.documents.get(id);
```

## Get typed metadata

Returns the document's registered typed metadata field values. A field is only included if its value is identical across every chunk of the document — this avoids showing a value that's mid-sync.

```ts
const metadata = await collection.documents.metadata(id);
```

## Rename a document

Sets the display `name` on a document. Useful for giving a human-readable label to a file ingested from a URI.

```ts
await collection.documents.rename(doc.id, "Q4 2024 Report");
```

## Update a document

Update `name`, `extraJson`, and/or typed `metadata` in one call. `metadata` keys that aren't registered fields on this collection (or whose value doesn't match the field's declared type) are silently dropped by the server — the request still succeeds, with no signal for which keys landed. Register fields first via the metadata-fields admin API.

```ts
await collection.documents.update(doc.id, {
  name: "Q4 2024 Report (final)",
  extraJson: JSON.stringify({ source: "zendesk" }),
  metadata: { reviewed: true },
});
```

| Option | Type | Description |
|---|---|---|
| `name` | `string \| undefined` | New display name |
| `extraJson` | `string \| undefined` | Replaces the document's freeform `extra_json` metadata |
| `metadata` | `Record<string, unknown> \| undefined` | Merges values into the document's typed metadata fields |

## Delete a document

Removes the document and all its chunks from the index. This is irreversible.

```ts
await collection.documents.delete(doc.id);
```
