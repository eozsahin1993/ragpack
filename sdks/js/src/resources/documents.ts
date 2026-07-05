import type { Document } from "../types.js";
import type { Requester } from "../requester.js";

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
   * Delete a document and all its chunks from this collection.
   * @param id - The document ID.
   */
  delete(id: string): Promise<void> {
    return this.req<void>(`/collections/${this.slug}/documents/${id}`, {
      method: "DELETE",
    });
  }
}
