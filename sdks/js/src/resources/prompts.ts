import type { Prompt } from "../types.js";
import type { Requester } from "../requester.js";

export class PromptsResource {
  constructor(private readonly req: Requester) {}

  /** List all prompt templates (built-in system prompts first, then custom). */
  async list(): Promise<Prompt[]> {
    const r = await this.req<{ system: Prompt[]; user: Prompt[] }>("/prompts");
    return [...(r.system ?? []), ...(r.user ?? [])];
  }

  /** Fetch a prompt template by slug. */
  get(slug: string): Promise<Prompt> {
    return this.req<Prompt>(`/prompts/${slug}`);
  }
}
