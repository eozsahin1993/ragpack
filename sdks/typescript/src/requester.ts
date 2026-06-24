import type { RagPackConfig } from "./types.js";

export type Requester = <T>(path: string, init?: RequestInit) => Promise<T>;

export function createRequester(config: RagPackConfig): Requester {
  return async function req<T>(path: string, init?: RequestInit): Promise<T> {
    const isFormData = init?.body instanceof FormData;
    const res = await fetch(`${config.baseUrl}/api/v1${path}`, {
      ...init,
      headers: {
        ...(!isFormData && { "Content-Type": "application/json" }),
        Authorization: `Bearer ${config.apiKey}`,
        ...init?.headers,
      },
    });

    if (res.status === 204) return undefined as T;

    const body = await res.json() as { error?: string };
    if (!res.ok) throw new RagPackError(res.status, body.error ?? res.statusText);
    return body as unknown as T;
  };
}

export class RagPackError extends Error {
  constructor(
    public readonly status: number,
    message: string
  ) {
    super(message);
    this.name = "RagPackError";
  }
}
