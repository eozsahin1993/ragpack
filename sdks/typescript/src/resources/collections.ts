import type { Collection } from "../types.js";
import type { Requester } from "../requester.js";
import { CollectionClient } from "../collection-client.js";

export class CollectionsResource {
  constructor(private readonly req: Requester) {}

  /** List all collections. */
  async list(): Promise<Collection[]> {
    const r = await this.req<{ collections: Collection[] }>("/collections");
    return r.collections;
  }

  /**
   * Create a new collection and return a scoped client for it.
   * @param name - Display name for the collection.
   * @param options.embedModel - Embedding model to use. Defaults to the server's configured provider.
   *
   * @example
   * ```ts
   * const collection = await client.collections.create("my-docs");
   * await collection.ingest(file);
   * ```
   */
  async create(name: string, options?: { embedModel?: string }): Promise<CollectionClient> {
    const col = await this.req<Collection>("/collections", {
      method: "POST",
      body: JSON.stringify({ name, embed_model: options?.embedModel }),
    });
    return new CollectionClient(col.slug, this.req);
  }

  /**
   * Get a collection by its slug.
   * @param slug - The collection's URL-safe identifier.
   */
  get(slug: string): Promise<Collection> {
    return this.req<Collection>(`/collections/${slug}`);
  }

  /**
   * Delete a collection and all its documents.
   * @param slug - The collection's URL-safe identifier.
   */
  delete(slug: string): Promise<void> {
    return this.req<void>(`/collections/${slug}`, { method: "DELETE" });
  }
}
