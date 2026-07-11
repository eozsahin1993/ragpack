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
  display_name?: string;
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
  name?: string;
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
  /** Cosine similarity score between 0 and 100. Higher is more relevant. */
  vector_similarity: number;
  /** Raw BM25 score from the keyword channel; present only for hybrid results. */
  keyword_bm25_score?: number;
  /** Weighted RRF fusion score, normalized so this batch's top result is 100. */
  rrf_score_normalized?: number;
  /** Raw RRF fusion score; only comparable within this query's own weights/k. */
  rrf_score?: number;
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
  chunk_header: string | null;
  /** The matched text chunk. */
  chunk_text: string | null;
  /** Optional metadata attached at ingest time. */
  extra_json: string | null;
  /** Typed metadata field values registered for this collection. */
  metadata?: Record<string, unknown>;
  vector_distance: number;
  /** Cosine similarity score between 0 and 100. Higher is more relevant. */
  vector_similarity: number;
  /** Raw BM25 score from the keyword channel; present only for hybrid results. */
  keyword_bm25_score?: number;
  /** Weighted RRF fusion score, normalized so this batch's top result is 100. */
  rrf_score_normalized?: number;
  /** Raw RRF fusion score; only comparable within this query's own weights/k. */
  rrf_score?: number;
}

export type FilterValue = string | number | boolean;

/** Operators usable against a single field in a {@link FilterExpression}. */
export interface FilterOps {
  $eq?: FilterValue;
  $ne?: FilterValue;
  $gt?: FilterValue;
  $gte?: FilterValue;
  $lt?: FilterValue;
  $lte?: FilterValue;
  $in?: FilterValue[];
  $nin?: FilterValue[];
  $exists?: boolean;
  /** str/timestamp fields only. */
  $like?: string;
  /** str fields only. */
  $ilike?: string;
  /** arr fields only. */
  $contains?: string;
  /** arr fields only. */
  $containsAny?: string[];
  /** arr fields only. */
  $containsAll?: string[];
}

/**
 * MongoDB-style filter expression, compiled server-side into a predicate over
 * a document's built-in and registered metadata fields.
 *
 * @example
 * ```ts
 * { mime_type: "application/pdf", score: { $gte: 4 } }
 * { $or: [{ tags: { $contains: "urgent" } }, { created_at: { $gte: "7 days ago" } }] }
 * ```
 */
export type FilterExpression =
  | { $and: FilterExpression[] }
  | { $or: FilterExpression[] }
  | { [field: string]: FilterValue | FilterOps };

/**
 * Per-request override of the weighted RRF merge between vector and keyword
 * search. Unset fields fall back to server defaults (semantic-favored 7:3).
 */
export interface HybridSettings {
  fullTextWeight?: number;
  semanticWeight?: number;
  rrfK?: number;
  /** Caps FTS candidates considered before fusion; FTS search has no native limit. */
  fullTextLimit?: number;
}
