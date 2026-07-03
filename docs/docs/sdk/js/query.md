---
sidebar_position: 4
---

# Query

RagPack has two query modes: **semantic search**, which returns ranked chunks, and **RAG**, which retrieves chunks and uses an LLM to produce a grounded answer.

## Semantic search

`findSimilar` embeds your query and returns the most relevant chunks from the collection, ranked by similarity score.

```ts
const results = await collection.findSimilar({
  query: "how do I configure authentication?",
  topK: 5,
});

for (const r of results) {
  console.log(r.similarity, r.chunkText);
}
```

| Option | Type | Description |
|---|---|---|
| `query` | `string` | Natural language search query |
| `topK` | `number` | Number of results to return. Defaults to `5`, max `100`. |

Each result contains:

| Field | Type | Description |
|---|---|---|
| `chunkText` | `string \| null` | The matched text |
| `fileUri` | `string` | Source document URI |
| `mimeType` | `string` | MIME type of the source document |
| `chunkIndex` | `number` | Position of this chunk within the document |
| `similarity` | `number` | Cosine similarity score (0–1). Higher is more relevant. |
| `distance` | `number` | Raw vector distance. Lower is more similar. |

## RAG (retrieve + generate)

`rag` runs the full pipeline server-side: it retrieves the most relevant chunks, builds context, fills in your prompt template, and calls the configured LLM — returning both the answer and the chunks used to produce it.

```ts
const { answer, chunks } = await collection.rag({
  query: "How do I reset my password?",
  topK: 5,
  promptSlug: "basic-rag",
  model: "gpt-4o",
  minSimilarity: 60,
});

console.log(answer);
```

| Option | Type | Description |
|---|---|---|
| `query` | `string` | The user's question |
| `topK` | `number` | Number of chunks to retrieve. Defaults to `5`. |
| `promptSlug` | `string` | Slug of the prompt template to use. Defaults to `"basic_rag"`. |
| `model` | `string` | LLM model name (e.g. `"gpt-4o"`, `"llama3"`). Falls back to server default if omitted. |
| `minSimilarity` | `number` | Minimum similarity score (0–100) a chunk must meet to be included. Omit to include all top-K results. |

The response contains:

| Field | Type | Description |
|---|---|---|
| `answer` | `string` | LLM-generated answer |
| `chunks` | `RagChunk[]` | Chunks used to build the context |
| `formattedPrompt` | `string` | The fully expanded prompt sent to the LLM |
| `promptSlug` | `string` | The prompt template used |
