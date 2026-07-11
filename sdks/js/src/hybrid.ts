import type { FilterExpression, HybridSettings } from "./types.js";

export interface HybridRequestOptions {
  filters?: FilterExpression;
  vectorSearchOnly?: boolean;
  hybridSettings?: HybridSettings;
}

/** Converts camelCase hybrid/filter options into the snake_case wire fields shared by /query and /rag. */
export function buildHybridBody(options: HybridRequestOptions): Record<string, unknown> {
  const { filters, vectorSearchOnly, hybridSettings } = options;
  return {
    ...(filters !== undefined && { filters }),
    ...(vectorSearchOnly !== undefined && { vector_search_only: vectorSearchOnly }),
    ...(hybridSettings !== undefined && {
      hybrid_settings: {
        ...(hybridSettings.fullTextWeight !== undefined && {
          full_text_weight: hybridSettings.fullTextWeight,
        }),
        ...(hybridSettings.semanticWeight !== undefined && {
          semantic_weight: hybridSettings.semanticWeight,
        }),
        ...(hybridSettings.rrfK !== undefined && { rrf_k: hybridSettings.rrfK }),
        ...(hybridSettings.fullTextLimit !== undefined && {
          full_text_limit: hybridSettings.fullTextLimit,
        }),
      },
    }),
  };
}
