import type { Document } from "../types.js";
import type { Requester } from "../requester.js";

export interface UpdateDocumentOptions {
  /** New display name. */
  name?: string;
  /** Replaces the document's freeform extra_json metadata. */
  extraJson?: string;
  /** Merges values into the document's typed metadata fields. */
  metadata?: Record<string, unknown>;
}

export class DocumentsResource {
  constructor(
    private readonly req: Requester,
    private readonly slug?: string
  ) {}

  /**
   * List documents in this collection.
   * @param options.limit - Max number of documents to return. Defaults to 50.
   * @param options.offset - Pagination offset. Defaults to 0.
   */
  async list(options?: { limit?: number; offset?: number }): Promise<Document[]> {
    const params = new URLSearchParams({
      limit: String(options?.limit ?? 50),
      offset: String(options?.offset ?? 0),
    });
    const r = await this.req<{ documents: Document[] }>(
      `/collections/${this.slug}/documents?${params}`
    );
    return r.documents;
  }

  /**
   * Get a single document by ID.
   * @param id - The document ID.
   */
  get(id: string): Promise<Document> {
    return this.req<Document>(`/collections/${this.slug}/documents/${id}`);
  }

  /**
   * Get this document's typed metadata field values.
   * A field is only included if its value is identical across every chunk
   * of the document — this avoids showing a value that's mid-sync.
   * @param id - The document ID.
   */
  async metadata(id: string): Promise<Record<string, unknown>> {
    const r = await this.req<{ metadata: Record<string, unknown> }>(
      `/collections/${this.slug}/documents/${id}/metadata`
    );
    return r.metadata;
  }

  /**
   * Rename a document.
   * @param id - The document ID.
   * @param name - The new display name.
   */
  rename(id: string, name: string): Promise<Document> {
    return this.req<Document>(`/collections/${this.slug}/documents/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ name }),
    });
  }

  /**
   * Update a document's name, extra_json, and/or typed metadata.
   * `metadata` keys that aren't registered fields on this collection (or whose
   * value doesn't match the field's declared type) are silently dropped by
   * the server — the request still succeeds, with no signal for which keys
   * landed. Register fields first via the metadata-fields admin API.
   * @param id - The document ID.
   *
   * @example
   * ```ts
   * await collection.documents.update(id, {
   *   metadata: { reviewed: true },
   *   extraJson: JSON.stringify({ source: "zendesk" }),
   * });
   * ```
   */
  update(id: string, options: UpdateDocumentOptions): Promise<Document> {
    return this.req<Document>(`/collections/${this.slug}/documents/${id}`, {
      method: "PATCH",
      body: JSON.stringify({
        ...(options.name !== undefined && { name: options.name }),
        ...(options.extraJson !== undefined && { extra_json: options.extraJson }),
        ...(options.metadata !== undefined && { metadata: options.metadata }),
      }),
    });
  }

  /**
   * Delete a document and all its chunks from this collection.
   * @param id - The document ID.
   */
  delete(id: string): Promise<void> {
    return this.req<void>(`/collections/${this.slug}/documents/${id}`, {
      method: "DELETE",
    });
  }
}
