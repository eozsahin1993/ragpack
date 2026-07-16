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
		content: `You are a helpful assistant. Answer the user's question using the information in the context below, including reasonable inferences that follow directly from what is stated — the answer does not need to be a verbatim quote. Answer naturally and directly — do not cite sources or reference the documents.

<context>
{{context}}
</context>

Only say "I don't have enough information to answer this." if the context truly has nothing relevant to the question. Do not explain why, do not reference the documents, and do not use outside knowledge.

Question: {{question}}`,
	},
	{
		id:   "a0000000-0000-0000-0000-000000000002",
		name: "RAG with Citations",
		slug: "rag_with_citations",
		content: `You are a helpful research assistant. Answer the question below using the provided context, including reasonable inferences that follow directly from what is stated.

<context>
{{context}}
</context>

Guidelines:
- Write a clear, direct answer. Cite sources the way an academic paper does: a bracketed number immediately after the claim it supports, e.g. "Growth elasticity of poverty depends on the level of inequality [1]." Use the number from the source's context label (e.g. [Source 2: ...] becomes [2]). If a claim draws on more than one source, cite them together, e.g. [1][2].
- Do not write "According to Source X" as a sentence prefix — the bracketed number is the citation.
- Only say "The provided documents don't contain enough information to answer this fully." if the context truly has nothing relevant to the question.
- Do not speculate or use outside knowledge.

Question: {{question}}`,
	},
	{
		id:   "a0000000-0000-0000-0000-000000000003",
		name: "Concise Answer",
		slug: "concise_rag",
		content: `Answer in 2-3 sentences using the context provided, including reasonable inferences that follow directly from what is stated. Answer naturally — do not cite sources. Only say "I don't have that information." if the context truly has nothing relevant to the question.

<context>
{{context}}
</context>

Question: {{question}}`,
	},
}
