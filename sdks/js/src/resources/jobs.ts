import { RagPackError } from "../requester.js";
import type { Job } from "../types.js";
import type { Requester } from "../requester.js";

export interface WaitOptions {
  /** How often to poll for job status, in milliseconds. Defaults to 1500. */
  pollIntervalMs?: number;
  /** Maximum time to wait before throwing, in milliseconds. Defaults to 300000 (5 min). */
  timeoutMs?: number;
}

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

  /**
   * Poll until the job reaches a terminal state (`complete` or `failed`).
   * Throws {@link RagPackError} if the job fails or the timeout is exceeded.
   *
   * @example
   * ```ts
   * const job = await collection.ingest(file);
   * await collection.jobs.waitUntilComplete(job.id);
   * console.log("ingestion done, ready to search");
   * ```
   */
  async waitUntilComplete(id: string, options: WaitOptions = {}): Promise<Job> {
    const { pollIntervalMs = 1500, timeoutMs = 300_000 } = options;
    const deadline = Date.now() + timeoutMs;

    while (Date.now() < deadline) {
      const job = await this.get(id);
      if (job.status === "complete") return job;
      if (job.status === "failed") {
        throw new RagPackError(422, job.error ?? "ingestion job failed");
      }
      await new Promise<void>((resolve) => setTimeout(resolve, pollIntervalMs));
    }

    throw new RagPackError(408, `job ${id} did not complete within ${timeoutMs}ms`);
  }
}
