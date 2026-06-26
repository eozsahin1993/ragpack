import { createRequester, RagPackError } from "./requester.js";
import { CollectionsResource } from "./resources/collections.js";
import { PromptsResource } from "./resources/prompts.js";
import { LLMsResource } from "./resources/llms.js";
import { EmbeddersResource } from "./resources/embedders.js";
import { CollectionClient } from "./collection-client.js";
import type { RagPackConfig } from "./types.js";

export { RagPackError };
export type { RagPackConfig, Collection, Job, Document, QueryResult, Prompt, RagChunk, RagResult } from "./types.js";
export type { FindSimilarOptions, IngestUriOptions, RagOptions } from "./collection-client.js";
export type { LLMInfo } from "./resources/llms.js";
export type { EmbedderInfo } from "./resources/embedders.js";

/**
 * RagPack client for interacting with a self-hosted RagPack RAG engine.
 *
 * @example
 * ```ts
 * const client = new RagPack({ baseUrl: "http://localhost:9000", apiKey: "rp_..." });
 *
 * // Manage collections
 * const col = await client.collections.create("my-docs");
 *
 * // Scope to a collection for all operations
 * const collection = client.collection(col.slug);
 * await collection.ingest(file);
 * const results = await collection.findSimilar({ query: "what is RagPack?" });
 *
 * // Discover configured providers
 * const { models, default: defaultModel } = await client.llms.list();
 * ```
 */
export class RagPack {
  /** Create and manage collections. */
  readonly collections: CollectionsResource;
  /** Fetch and expand prompt templates. */
  readonly prompts: PromptsResource;
  /** List configured LLM models. */
  readonly llms: LLMsResource;
  /** List configured embedding models. */
  readonly embedders: EmbeddersResource;

  private readonly _req: ReturnType<typeof createRequester>;

  constructor(config: RagPackConfig) {
    this._req = createRequester(config);
    this.collections = new CollectionsResource(this._req);
    this.prompts = new PromptsResource(this._req);
    this.llms = new LLMsResource(this._req);
    this.embedders = new EmbeddersResource(this._req);
  }

  /**
   * Scope all operations to a specific collection.
   * @param slug - The collection's URL-safe identifier.
   *
   * @example
   * ```ts
   * const collection = client.collection("my-docs");
   * await collection.ingest(file);
   * await collection.findSimilar({ query: "..." });
   * await collection.jobs.list();
   * ```
   */
  collection(slug: string): CollectionClient {
    return new CollectionClient(slug, this._req);
  }
}
