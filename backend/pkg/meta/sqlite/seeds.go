package sqlite

// System prompts are defined here as Go constants. Edit and rebuild to update them —
// no migration needed. The store upserts these on every startup.

type systemPromptSeed struct {
	id      string
	name    string
	slug    string
	content string
}

var systemPrompts = []systemPromptSeed{
	{
		id:   "a0000000-0000-0000-0000-000000000001",
		name: "Basic RAG",
		slug: "basic_rag",
		content: `You are a helpful assistant. Answer the user's question using only the information in the context below. Answer naturally and directly — do not cite sources or reference the documents.

<context>
{{context}}
</context>

If the context does not contain the answer, say only: "I don't have enough information to answer this." Do not explain why, do not reference the documents, and do not use outside knowledge.

Question: {{question}}`,
	},
	{
		id:   "a0000000-0000-0000-0000-000000000002",
		name: "RAG with Citations",
		slug: "rag_with_citations",
		content: `You are a helpful research assistant. Answer the question below using only the provided context.

<context>
{{context}}
</context>

Guidelines:
- Cite sources inline where relevant (e.g. "According to [Source 1]...").
- If the context is insufficient, say: "The provided documents don't contain enough information to answer this fully."
- Do not speculate or use outside knowledge.

Question: {{question}}`,
	},
	{
		id:   "a0000000-0000-0000-0000-000000000003",
		name: "Concise Answer",
		slug: "concise_rag",
		content: `Answer in 2-3 sentences using only the context provided. Answer naturally — do not cite sources. If the answer is not in the context, say "I don't have that information."

<context>
{{context}}
</context>

Question: {{question}}`,
	},
}
