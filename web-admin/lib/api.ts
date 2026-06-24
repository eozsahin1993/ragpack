const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:9000";

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  const body = await res.json();
  if (!res.ok) throw new Error(body.error ?? res.statusText);
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

export interface QueryResultItem {
  source: string;
  file_uri: string;
  mime_type: string;
  chunk_index: number;
  chunk_text: string | null;
  extra_json: string | null;
  distance: number;
  similarity: number;
}

export const api = {
  collections: {
    list: () => req<{ collections: Collection[] }>("/api/v1/collections"),
    create: (body: { name: string; embed_model: string }) =>
      req<Collection>("/api/v1/collections", { method: "POST", body: JSON.stringify(body) }),
    get: (slug: string) => req<Collection>(`/api/v1/collections/${slug}`),
    delete: (slug: string) =>
      req<void>(`/api/v1/collections/${slug}`, { method: "DELETE" }),
  },
  jobs: {
    all: () => req<{ jobs: Job[] }>("/api/v1/jobs"),
    byCollection: (slug: string) =>
      req<{ jobs: Job[] }>(`/api/v1/collections/${slug}/jobs`),
    get: (slug: string, id: string) =>
      req<{ job: Job }>(`/api/v1/collections/${slug}/jobs/${id}`),
  },
  ingest: {
    uri: (slug: string, body: { file_uri: string; mime_type: string }) =>
      req<Job>(`/api/v1/collections/${slug}/ingest`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    upload: async (slug: string, file: File): Promise<Job> => {
      const form = new FormData();
      form.append("file", file);
      const res = await fetch(`${BASE}/api/v1/collections/${slug}/ingest`, {
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
        `/api/v1/collections/${slug}/documents?limit=${limit}&offset=${offset}`
      ),
    delete: (slug: string, id: string) =>
      req<void>(`/api/v1/collections/${slug}/documents/${id}`, { method: "DELETE" }),
  },
  query: {
    run: (slug: string, body: { query: string; top_k: number }) =>
      req<{ results: QueryResultItem[] }>(`/api/v1/collections/${slug}/query`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
  },
};
