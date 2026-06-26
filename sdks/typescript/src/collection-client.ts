import type { Requester } from "./requester.js";
import type { Job, QueryResult, RagResult } from "./types.js";
import { JobsResource } from "./resources/jobs.js";
import { DocumentsResource } from "./resources/documents.js";
import { QueryResource } from "./resources/query.js";

export interface RagOptions {
  /** Slug of the prompt template to use. */
  promptSlug: string;
  /** The user's question. Substituted into `{{question}}`. */
  query: string;
  /** Number of chunks to retrieve. Defaults to 5. */
  topK?: number;
  /** LLM model name to use (e.g. `"gpt-4o"`, `"claude-opus-4-8"`). Falls back to server default. */
  model?: string;
  /** Minimum similarity score (0–100) a chunk must meet to be included in context. Omit to include all top_k results. */
  minSimilarity?: number;
}

export interface FindSimilarOptions {
  /** The search query text. */
  query: string;
  /** Number of results to return. Defaults to 5. */
  topK?: number;
}

export interface IngestUriOptions {
  /** Remote file URI (e.g. `s3://bucket/file.pdf`). */
  uri: string;
  /** MIME type of the file. Detected from the file extension if omitted. */
  mimeType?: string;
}

/**
 * A collection-scoped client. All operations target the collection
 * this was created with.
 *
 * @example
 * ```ts
 * const collection = client.collection("my-docs");
 *
 * await collection.ingest(file);
 * await collection.ingest({ uri: "s3://bucket/file.pdf", mimeType: "application/pdf" });
 * const results = await collection.findSimilar({ query: "how does auth work?" });
 * ```
 */
export class CollectionClient {
  /** Inspect ingestion jobs for this collection. */
  readonly jobs: JobsResource;
  /** Manage indexed documents in this collection. */
  readonly documents: DocumentsResource;

  private readonly _req: Requester;
  private readonly _query: QueryResource;

  constructor(
    private readonly slug: string,
    req: Requester
  ) {
    this._req = req;
    this.jobs = new JobsResource(req, slug);
    this.documents = new DocumentsResource(req, slug);
    this._query = new QueryResource(req, slug);
  }

  /**
   * Ingest a document into this collection.
   *
   * Pass a `File` or `Blob` to upload directly, or an object with `uri` and
   * `mimeType` to ingest from a remote URI.
   *
   * @example
   * ```ts
   * // File upload
   * await collection.ingest(file);
   *
   * // Remote URI
   * await collection.ingest({ uri: "s3://bucket/report.pdf", mimeType: "application/pdf" });
   * ```
   */
  ingest(file: File | Blob, filename?: string): Promise<Job>;
  ingest(options: IngestUriOptions): Promise<Job>;
  ingest(input: File | Blob | IngestUriOptions, filename?: string): Promise<Job> {
    if (input instanceof Blob) {
      const form = new FormData();
      form.append("file", input, filename);
      return this._req<Job>(`/collections/${this.slug}/ingest`, {
        method: "POST",
        body: form,
      });
    }
    return this._req<Job>(`/collections/${this.slug}/ingest`, {
      method: "POST",
      body: JSON.stringify({ file_uri: input.uri, mime_type: input.mimeType }),
    });
  }

  /**
   * Find semantically similar content in this collection.
   * @example
   * ```ts
   * const results = await collection.findSimilar({ query: "what is RagPack?", topK: 10 });
   * ```
   */
  findSimilar(options: FindSimilarOptions): Promise<QueryResult[]> {
    return this._query.findSimilar(options);
  }

  /**
   * Full RAG pipeline: retrieve relevant chunks, build context, expand the
   * prompt template, and call the configured LLM — all server-side.
   *
   * @example
   * ```ts
   * const { answer, chunks } = await collection.rag({
   *   prompt: "basic-rag",
   *   query: "How do I reset my password?",
   * });
   * ```
   */
  rag(options: RagOptions): Promise<RagResult> {
    return this._req<RagResult>(`/collections/${this.slug}/rag`, {
      method: "POST",
      body: JSON.stringify({
        query: options.query,
        top_k: options.topK ?? 5,
        prompt_slug: options.promptSlug,
        ...(options.model ? { model: options.model } : {}),
        ...(options.minSimilarity != null ? { min_similarity: options.minSimilarity } : {}),
      }),
    });
  }
}
