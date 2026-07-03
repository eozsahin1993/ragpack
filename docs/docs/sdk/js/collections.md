---
sidebar_position: 2
---

# Collections

A collection is a named bucket for your documents. Each collection has its own vector index and can be queried independently. All documents you ingest belong to exactly one collection.

Collections are identified by a `slug` — a URL-safe version of the name generated automatically on creation (e.g. `"My Docs"` → `"my-docs"`).

## Create a collection

```ts
const collection = await client.collections.create("My Docs");
```

`create` returns a scoped `CollectionClient` ready to use immediately.

You can optionally override the embedding model and chunking behaviour for this collection:

```ts
const collection = await client.collections.create("My Docs", {
  embedModel: "text-embedding-3-small",
  chunkConfig: {
    strategy: "section",
    size: 1500,
    overlap: 150,
  },
});
```

| Option | Type | Description |
|---|---|---|
| `embedModel` | `string` | Embedding model to use. Defaults to the server's configured provider. |
| `chunkConfig.strategy` | `string` | Chunking strategy. Defaults to `"auto"`. See below. |
| `chunkConfig.size` | `number` | Max characters per chunk. |
| `chunkConfig.overlap` | `number` | Characters of overlap carried into the next chunk for context continuity. |

:::note
The embedding model is fixed at creation time and cannot be changed afterwards. All documents in a collection must be embedded with the same model so that query vectors and document vectors are comparable. To switch models, create a new collection and re-ingest your documents.
:::

### Chunking strategies

The strategy controls how documents are split into chunks before embedding. The default is `"auto"`, which picks the best strategy based on the file's MIME type — this is the right choice for most collections that ingest a mix of file types.

| Strategy | How it works |
|---|---|
| `"auto"` | Delegates to a per-MIME default. Picks the strategy below that best fits each file type automatically. |
| `"section"` | Splits on headings. Each heading and its content become one chunk. Oversized sections fall back to sliding window. |
| `"paragraph"` | Groups blank-line-separated paragraphs up to `size` characters, carrying `overlap` characters into the next group. |
| `"sliding_window"` | Fixed-size rolling window over raw text with `overlap` characters of carry-over between windows. |
| `"unit"` | One chunk per logical unit (e.g. one slide). Oversized units are split further. |
| `"row_group"` | Groups rows up to `size` characters, prepending the header row to each group so every chunk is self-contained. |

**Auto strategy defaults by file type:**

| File type | Strategy |
|---|---|
| `.md`, `.html` | `section` |
| `.txt`, `.docx` | `paragraph` |
| `.pdf` | `sliding_window` |
| `.pptx` | `unit` |
| `.xlsx`, `.csv`, `.json`, `.xml` | `row_group` |

## List all collections

```ts
const collections = await client.collections.list();

for (const col of collections) {
  console.log(col.slug, col.name, col.embed_model);
}
```

Each collection object contains:

| Field | Type | Description |
|---|---|---|
| `id` | `string` | Unique identifier |
| `name` | `string` | Display name |
| `slug` | `string` | URL-safe identifier used in API calls |
| `embed_model` | `string` | Embedding model used for this collection |
| `vector_dim` | `number` | Embedding dimension |
| `created_at` | `string` | ISO 8601 timestamp |
| `chunk_config` | `object` | Present only when chunking defaults are overridden |

## Get a collection

```ts
const col = await client.collections.get("my-docs");
```

## Delete a collection

Deletes the collection and all its documents and chunks. This is irreversible.

```ts
await client.collections.delete("my-docs");
```
