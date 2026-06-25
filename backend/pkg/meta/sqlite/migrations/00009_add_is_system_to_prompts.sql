-- +goose Up
ALTER TABLE prompts ADD COLUMN is_system INTEGER NOT NULL DEFAULT 0;

INSERT INTO prompts (id, name, slug, content, is_system, created_at, updated_at) VALUES
(
  'a0000000-0000-0000-0000-000000000001',
  'Basic RAG',
  'basic_rag',
  'You are a helpful assistant. Answer the question using only the provided context. If the answer is not found in the context, say "I don''t have enough information to answer that."

Context:
{{context}}

Question: {{question}}

Answer:',
  1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
),
(
  'a0000000-0000-0000-0000-000000000002',
  'RAG with Sections',
  'rag_with_sections',
  'You are a helpful assistant with access to structured documents. Each context block shows the section it came from so you can cite your sources accurately.

Answer the question using only the provided context. If the answer is not found in the context, say "I don''t have enough information to answer that."

Context:
{{context}}

Question: {{question}}

Answer:',
  1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
),
(
  'a0000000-0000-0000-0000-000000000003',
  'Concise RAG',
  'concise_rag',
  'Answer the question in 2-3 sentences using only the context provided. If the answer cannot be found in the context, respond with "I don''t know."

Context:
{{context}}

Question: {{question}}',
  1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
);

-- +goose Down
DELETE FROM prompts WHERE is_system = 1;
ALTER TABLE prompts DROP COLUMN is_system;
