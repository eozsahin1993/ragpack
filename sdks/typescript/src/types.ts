/**
 * Controls how ingested documents are split into chunks before embedding.
 *
 * - `"auto"` — picks the best strategy based on the file's MIME type (default).
 * - `"paragraph"` — splits on blank lines; good for prose (plain text, DOCX).
 * - `"sliding_window"` — fixed-size rolling window with overlap; good for dense text like PDFs.
 * - `"section"` — splits on headings, preserving section context; good for Markdown and HTML.
 * - `"unit"` — one chunk per logical unit (e.g. one slide for PPTX); good for structured presentations.
 * - `"row_group"` — groups rows with a repeated header row per chunk; good for spreadsheets (XLSX).
 */
export type ChunkStrategy =
  | "auto"
  | "paragraph"
  | "sliding_window"
  | "section"
  | "unit"
  | "row_group";

export interface RagPackConfig {
  /** Base URL of your RagPack backend (e.g. `http://localhost:9000`). */
  baseUrl: string;
  /** API key generated on first backend startup. */
  apiKey: string;
}

/** Per-collection chunking overrides. Omitted when all fields use server defaults. */
export interface ChunkConfig {
  strategy?: ChunkStrategy;
  /** Max characters per chunk. */
  size?: number;
  /** Characters of overlap carried into the next chunk for context continuity. */
  overlap?: number;
}

export interface Collection {
  id: string;
  name: string;
  /** URL-safe identifier used in API calls. */
  slug: string;
  embed_model: string;
  vector_dim: number;
  created_at: string;
  /** Present only when this collection overrides the server chunking defaults. */
  chunk_config?: ChunkConfig;
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

export interface Prompt {
  id: string;
  name: string;
  slug: string;
  /** Template content. Use `{{context}}` and `{{question}}` as placeholders. */
  content: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface RagChunk {
  source: string;
  file_uri: string;
  chunk_index: number;
  chunk_header: string | null;
  chunk_text: string | null;
  similarity: number;
}

export interface RagResult {
  formatted_prompt: string;
  answer: string;
  chunks: RagChunk[];
  prompt_slug: string;
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
