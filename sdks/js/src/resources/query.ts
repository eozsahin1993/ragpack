import type { QueryResult } from "../types.js";
import type { Requester } from "../requester.js";

export interface FindSimilarOptions {
  /** The search query text. */
  query: string;
  /** Number of results to return. Defaults to 5. */
  topK?: number;
}

export class QueryResource {
  constructor(
    private readonly req: Requester,
    private readonly slug?: string
  ) {}

  /**
   * Find semantically similar content within this collection.
   * @returns Ranked results with similarity scores and source chunks.
   */
  async findSimilar(options: FindSimilarOptions): Promise<QueryResult[]> {
    const r = await this.req<{ results: QueryResult[] }>(
      `/collections/${this.slug}/query`,
      {
        method: "POST",
        body: JSON.stringify({ query: options.query, top_k: options.topK ?? 5 }),
      }
    );
    return r.results;
  }
}
