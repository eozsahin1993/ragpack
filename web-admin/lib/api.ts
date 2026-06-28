async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  const text = await res.text();
  const body = text ? JSON.parse(text) : null;
  if (!res.ok) throw new Error(body?.error ?? res.statusText);
  return body as T;
}

export interface Collection {
  id: string;
  name: string;
  slug: string;
  embed_model: string;
  vector_dim: number;
  table_name: string;
  created_at: string;
}

export interface Job {
  id: string;
  collection_id: string;
  file_uri: string;
  mime_type: string;
  status: string;
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface Document {
  id: string;
  collection_id: string;
  job_id: string;
  file_uri: string;
  mime_type: string;
  external_id?: string;
  extra_json?: string;
  chunk_count: number;
  status: "ingesting" | "complete" | "failed";
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface Chunk {
  id: string;
  document_id: string;
  chunk_hash: string;
  chunk_index: number;
  mime_type: string;
  file_uri: string;
  chunk_text: string | null;
  chunk_header: string | null;
  created_at: string;
}

export interface QueryResultItem {
  source: string;
  file_uri: string;
  mime_type: string;
  chunk_index: number;
  chunk_header: string | null;
  chunk_text: string | null;
  extra_json: string | null;
  distance: number;
  similarity: number;
}

export interface RagChunk {
  source: string;
  file_uri: string;
  chunk_index: number;
  chunk_header: string | null;
  chunk_text: string | null;
  similarity: number;
}

export interface RagResponse {
  formatted_prompt: string;
  answer: string;
  chunks: RagChunk[];
  prompt_slug: string;
}

export interface HealthInfo {
  status: string;
  version: string;
  uptime: string;
}

export interface EmbedderInfo {
  models: string[];
  default: string;
}

export interface LlmInfo {
  models: string[];
  default: string;
}

export interface Prompt {
  id: string;
  name: string;
  slug: string;
  content: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export const api = {
  health: {
    get: () => req<HealthInfo>("/admin/health"),
  },
  embedders: {
    list: () => req<EmbedderInfo>("/admin/embeddings"),
  },
  collections: {
    list: () => req<{ collections: Collection[] }>("/admin/collections"),
    create: (body: { name: string; embed_model?: string }) =>
      req<Collection>("/admin/collections", { method: "POST", body: JSON.stringify(body) }),
    get: (slug: string) => req<Collection>(`/admin/collections/${slug}`),
    delete: (slug: string) =>
      req<void>(`/admin/collections/${slug}`, { method: "DELETE" }),
  },
  jobs: {
    all: () => req<{ jobs: Job[] }>("/admin/jobs"),
    byCollection: (slug: string) =>
      req<{ jobs: Job[] }>(`/admin/collections/${slug}/jobs`),
    get: (slug: string, id: string) =>
      req<{ job: Job }>(`/admin/collections/${slug}/jobs/${id}`),
  },
  ingest: {
    uri: (slug: string, body: { file_uri: string; mime_type: string }) =>
      req<Job>(`/admin/collections/${slug}/ingest`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    refresh: (slug: string, body: { file_uri: string; mime_type: string }) =>
      req<Job>(`/admin/collections/${slug}/ingest?refresh=true`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    upload: async (slug: string, file: File): Promise<Job> => {
      const form = new FormData();
      form.append("file", file);
      const res = await fetch(`/admin/collections/${slug}/ingest`, {
        method: "POST",
        body: form,
      });
      const body = await res.json();
      if (!res.ok) throw new Error(body.error ?? res.statusText);
      return body as Job;
    },
  },
  documents: {
    list: (slug: string, limit = 50, offset = 0) =>
      req<{ documents: Document[]; total: number; limit: number; offset: number }>(
        `/admin/collections/${slug}/documents?limit=${limit}&offset=${offset}`
      ),
    get: (slug: string, id: string) =>
      req<Document>(`/admin/collections/${slug}/documents/${id}`),
    chunks: (slug: string, id: string) =>
      req<{ chunks: Chunk[]; total: number }>(`/admin/collections/${slug}/documents/${id}/chunks`),
    delete: (slug: string, id: string) =>
      req<void>(`/admin/collections/${slug}/documents/${id}`, { method: "DELETE" }),
  },
  llms: {
    list: () => req<LlmInfo>("/admin/llms"),
  },
  query: {
    run: (slug: string, body: { query: string; top_k: number }) =>
      req<{ results: QueryResultItem[] }>(`/admin/collections/${slug}/query`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    rag: (slug: string, body: { query: string; top_k: number; prompt_slug: string; model?: string; min_similarity?: number }) =>
      req<RagResponse>(`/admin/collections/${slug}/rag`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
  },
  prompts: {
    list: () => req<{ system: Prompt[]; user: Prompt[]; total: number }>("/admin/prompts"),
    create: (body: { name: string; content: string }) =>
      req<Prompt>("/admin/prompts", { method: "POST", body: JSON.stringify(body) }),
    get: (slug: string) => req<Prompt>(`/admin/prompts/${slug}`),
    update: (slug: string, body: { name?: string; content?: string }) =>
      req<Prompt>(`/admin/prompts/${slug}`, { method: "PATCH", body: JSON.stringify(body) }),
    delete: (slug: string) =>
      req<void>(`/admin/prompts/${slug}`, { method: "DELETE" }),
  },
};
