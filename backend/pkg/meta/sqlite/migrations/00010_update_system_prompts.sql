-- +goose Up
UPDATE prompts SET
  name = 'Basic RAG',
  content = 'You are a helpful assistant. Answer the user''s question using only the information in the context below.

<context>
{{context}}
</context>

If the context does not contain the answer, respond: "I don''t have enough information in the provided documents to answer this." Do not use outside knowledge to fill gaps.

Question: {{question}}',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'basic_rag' AND is_system = 1;

UPDATE prompts SET
  name = 'RAG with Citations',
  slug = 'rag_with_citations',
  content = 'You are a helpful research assistant. Answer the question below using only the provided context.

<context>
{{context}}
</context>

Guidelines:
- Cite sources inline where relevant (e.g. "According to [Source 1]...").
- If the context is insufficient, say: "The provided documents don''t contain enough information to answer this fully."
- Do not speculate or use outside knowledge.

Question: {{question}}',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'rag_with_sections' AND is_system = 1;

UPDATE prompts SET
  name = 'Concise Answer',
  content = 'Answer in 2-3 sentences using only the context provided. If the answer is not in the context, say "I don''t have that information in the provided documents."

<context>
{{context}}
</context>

Question: {{question}}',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'concise_rag' AND is_system = 1;

-- +goose Down
UPDATE prompts SET
  name = 'Basic RAG',
  content = 'You are a helpful assistant. Answer the question using only the provided context. If the answer is not found in the context, say "I don''t have enough information to answer that."

Context:
{{context}}

Question: {{question}}

Answer:',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'basic_rag' AND is_system = 1;

UPDATE prompts SET
  name = 'RAG with Sections',
  slug = 'rag_with_sections',
  content = 'You are a helpful assistant with access to structured documents. Each context block shows the section it came from so you can cite your sources accurately.

Answer the question using only the provided context. If the answer is not found in the context, say "I don''t have enough information to answer that."

Context:
{{context}}

Question: {{question}}

Answer:',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'rag_with_citations' AND is_system = 1;

UPDATE prompts SET
  name = 'Concise RAG',
  content = 'Answer the question in 2-3 sentences using only the context provided. If the answer cannot be found in the context, respond with "I don''t know."

Context:
{{context}}

Question: {{question}}',
  updated_at = CURRENT_TIMESTAMP
WHERE slug = 'concise_rag' AND is_system = 1;
