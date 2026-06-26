import type { Requester } from "../requester.js";

export interface LLMInfo {
  models: string[];
  default: string;
}

export class LLMsResource {
  constructor(private readonly req: Requester) {}

  /** List all configured LLM models and the server default. */
  list(): Promise<LLMInfo> {
    return this.req<LLMInfo>("/llms");
  }
}
