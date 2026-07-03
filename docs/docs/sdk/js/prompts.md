---
sidebar_position: 5
---

# Prompts

Prompts are templates used by the RAG endpoint to turn retrieved chunks into an LLM answer. Each prompt has a `{{context}}` placeholder where the retrieved chunks are inserted, and a `{{question}}` placeholder for the user's query.

RagPack ships with three built-in system prompts. You can also create your own via the admin UI.

## Built-in prompts

| Slug | Name | Description |
|---|---|---|
| `basic_rag` | Basic RAG | Answers naturally using only the provided context. Does not cite sources. |
| `rag_with_citations` | RAG with Citations | Answers with inline source citations (e.g. "According to [Source 1]..."). |
| `concise_rag` | Concise Answer | Answers in 2–3 sentences. |

`basic_rag` is the default when no `promptSlug` is passed to `collection.rag()`.

## List prompts

```ts
const prompts = await client.prompts.list();

for (const p of prompts) {
  console.log(p.slug, p.name, p.isSystem);
}
```

System prompts (`isSystem: true`) are listed first, followed by any custom prompts you've created.

## Get a prompt

```ts
const prompt = await client.prompts.get("basic_rag");
console.log(prompt.content);
```

## Using a prompt in RAG

Pass the prompt's `slug` to `collection.rag()`:

```ts
const { answer } = await collection.rag({
  query: "What is the refund policy?",
  promptSlug: "rag_with_citations",
  model: "gpt-4o",
});
```

## Custom prompts

Custom prompts can be created and managed in the admin UI. Use `{{context}}` and `{{question}}` as placeholders in your template:

```
Answer the question below in a formal tone using only the provided context.

<context>
{{context}}
</context>

Question: {{question}}
```

Once created, use the prompt's slug in `collection.rag()` the same way as a built-in prompt.
