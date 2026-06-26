import type { Requester } from "../requester.js";

export interface EmbedderInfo {
  models: string[];
  default: string;
}

export class EmbeddersResource {
  constructor(private readonly req: Requester) {}

  /** List all configured embedding models and the server default. */
  list(): Promise<EmbedderInfo> {
    return this.req<EmbedderInfo>("/embedders");
  }
}
