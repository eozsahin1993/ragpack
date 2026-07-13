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
  display_name?: string;
  intent?: string;
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
  name?: string;
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

export interface MetadataField {
  id: string;
  collection_id: string;
  name: string;
  type: "str" | "num" | "bool" | "date" | "arr";
  slot: number;
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
  metadata?: Record<string, unknown>;
  vector_distance: number;
  vector_similarity: number;
  keyword_bm25_score?: number;
  rrf_score?: number;
  rrf_score_normalized?: number;
}

export interface HybridSettings {
  full_text_weight?: number;
  semantic_weight?: number;
  rrf_k?: number;
  full_text_limit?: number;
}

export interface RagChunk {
  source: string;
  file_uri: string;
  chunk_index: number;
  chunk_header: string | null;
  chunk_text: string | null;
  vector_similarity: number;
  keyword_bm25_score?: number;
  rrf_score?: number;
  rrf_score_normalized?: number;
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

export type Permission = "read" | "write" | "both";
export type AdminResourceType = "keys" | "prompts" | "collections" | "*";

export interface CollectionGrant {
  id: string;
  api_key_id: string;
  collection_id?: string; // absent = every collection (wildcard)
  permission: Permission;
  created_at: string;
}

export interface AdminGrant {
  id: string;
  api_key_id: string;
  resource_type: AdminResourceType;
  permission: Permission;
  created_at: string;
}

export interface ApiKey {
  id: string;
  name: string;
  key_hint: string;
  created_at: string;
  grants: CollectionGrant[];
  admin_grants?: AdminGrant[];
}

export interface CreateApiKeyGrant {
  collection_slug?: string; // omitted = every collection (wildcard)
  permission: Permission;
}

export interface CreateApiKeyAdminGrant {
  resource_type: AdminResourceType;
  permission: Permission;
}

export interface CreatedApiKey extends ApiKey {
  key: string; // plaintext — only ever present in the create response
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
    getById: (id: string) => req<Collection>(`/admin/collections/id/${id}`),
    delete: (slug: string) =>
      req<void>(`/admin/collections/${slug}`, { method: "DELETE" }),
  },
  jobs: {
    all: (limit = 50, offset = 0) =>
      req<{ jobs: Job[]; total: number; limit: number; offset: number }>(`/admin/jobs?limit=${limit}&offset=${offset}`),
    byCollection: (slug: string) =>
      req<{ jobs: Job[] }>(`/admin/collections/${slug}/jobs`),
    get: (slug: string, id: string) =>
      req<{ job: Job }>(`/admin/collections/${slug}/jobs/${id}`),
    delete: (id: string) =>
      req<void>(`/admin/jobs/${id}`, { method: "DELETE" }),
  },
  ingest: {
    uri: (slug: string, body: { file_uri: string; mime_type: string; extra_json?: string }) =>
      req<Job>(`/admin/collections/${slug}/ingest`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    refresh: (slug: string, body: { file_uri: string; mime_type: string; extra_json?: string }) =>
      req<Job>(`/admin/collections/${slug}/ingest?refresh=true`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    upload: async (slug: string, file: File, extraJSON?: string): Promise<Job> => {
      const form = new FormData();
      form.append("file", file);
      if (extraJSON) form.append("extra_json", extraJSON);
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
    all: (limit = 50, offset = 0, status?: Document["status"]) =>
      req<{ documents: Document[]; total: number; limit: number; offset: number }>(
        `/admin/documents?limit=${limit}&offset=${offset}${status ? `&status=${status}` : ""}`
      ),
    // slug is only needed to scope the URL under a collection; pass null to
    // hit the slug-less /admin/documents/:id route instead (same document,
    // looked up by id alone — see backend/pkg/api/documents/handler.go).
    get: (slug: string | null, id: string) =>
      req<Document>(slug ? `/admin/collections/${slug}/documents/${id}` : `/admin/documents/${id}`),
    chunks: (slug: string | null, id: string, limit = 20, offset = 0) =>
      req<{ chunks: Chunk[]; total: number; limit: number; offset: number }>(
        `${slug ? `/admin/collections/${slug}/documents/${id}/chunks` : `/admin/documents/${id}/chunks`}?limit=${limit}&offset=${offset}`
      ),
    update: (slug: string | null, id: string, body: { name?: string; extra_json?: string; metadata?: Record<string, unknown> }) =>
      req<Document>(
        slug ? `/admin/collections/${slug}/documents/${id}` : `/admin/documents/${id}`,
        { method: "PATCH", body: JSON.stringify(body) }
      ),
    metadata: (slug: string | null, id: string) =>
      req<{ metadata: Record<string, unknown> }>(
        slug ? `/admin/collections/${slug}/documents/${id}/metadata` : `/admin/documents/${id}/metadata`
      ),
    delete: (slug: string | null, id: string) =>
      req<void>(
        slug ? `/admin/collections/${slug}/documents/${id}` : `/admin/documents/${id}`,
        { method: "DELETE" }
      ),
  },
  llms: {
    list: () => req<LlmInfo>("/admin/llms"),
  },
  metadataFields: {
    list: (slug: string) =>
      req<{ fields: MetadataField[] }>(`/admin/collections/${slug}/metadata-fields`),
    register: (slug: string, fields: { name: string; type: string }[]) =>
      req<{ fields: MetadataField[] }>(`/admin/collections/${slug}/metadata-fields`, {
        method: "POST",
        body: JSON.stringify({ fields }),
      }),
    delete: (slug: string, fieldName: string) =>
      req<void>(`/admin/collections/${slug}/metadata-fields/${fieldName}`, { method: "DELETE" }),
  },
  query: {
    run: (slug: string, body: {
      query: string;
      top_k: number;
      filters?: unknown;
      vector_search_only?: boolean;
      hybrid_settings?: HybridSettings;
    }) =>
      req<{ results: QueryResultItem[] }>(`/admin/collections/${slug}/query`, {
        method: "POST",
        body: JSON.stringify(body),
      }),
    rag: (slug: string, body: {
      query: string;
      top_k: number;
      prompt_slug: string;
      model?: string;
      min_similarity?: number;
      filters?: unknown;
      vector_search_only?: boolean;
      hybrid_settings?: HybridSettings;
    }) =>
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
  keys: {
    list: () => req<{ keys: ApiKey[] }>("/admin/keys"),
    create: (body: { name: string; grants: CreateApiKeyGrant[]; admin_grants?: CreateApiKeyAdminGrant[] }) =>
      req<CreatedApiKey>("/admin/keys", { method: "POST", body: JSON.stringify(body) }),
    delete: (id: string) =>
      req<void>(`/admin/keys/${id}`, { method: "DELETE" }),
  },
};
