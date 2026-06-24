import type { Job } from "../types.js";
import type { Requester } from "../requester.js";

export class JobsResource {
  constructor(
    private readonly req: Requester,
    private readonly slug?: string
  ) {}

  /** List ingestion jobs for this collection, or all jobs if no collection is scoped. */
  async list(): Promise<Job[]> {
    const path = this.slug ? `/collections/${this.slug}/jobs` : "/jobs";
    const r = await this.req<{ jobs: Job[] }>(path);
    return r.jobs;
  }

  /**
   * Get a single ingestion job by ID.
   * @param id - The job ID.
   */
  async get(id: string): Promise<Job> {
    const r = await this.req<{ job: Job }>(`/collections/${this.slug}/jobs/${id}`);
    return r.job;
  }
}
