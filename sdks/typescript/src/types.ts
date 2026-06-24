export interface RagPackConfig {
  /** Base URL of your RagPack backend (e.g. `http://localhost:9000`). */
  baseUrl: string;
  /** API key generated on first backend startup. */
  apiKey: string;
}

export interface Collection {
  id: string;
  name: string;
  /** URL-safe identifier used in API calls. */
  slug: string;
  embed_model: string;
  vector_dim: number;
  created_at: string;
}

export interface Job {
  id: string;
  collection_id: string;
  file_uri: string;
  mime_type: string;
  status: "pending" | "processing" | "complete" | "failed";
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
  chunk_count: number;
  status: "ingesting" | "complete" | "failed";
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface QueryResult {
  source: string;
  file_uri: string;
  mime_type: string;
  chunk_index: number;
  /** The matched text chunk. */
  chunk_text: string | null;
  /** Optional metadata attached at ingest time. */
  extra_json: string | null;
  distance: number;
  /** Cosine similarity score between 0 and 1. Higher is more relevant. */
  similarity: number;
}
