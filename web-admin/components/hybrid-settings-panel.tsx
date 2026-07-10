"use client";

import { useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { HybridSettings } from "@/lib/api";

export interface HybridConfig {
  vectorSearchOnly: boolean;
  fullTextWeight: string;
  semanticWeight: string;
  rrfK: string;
  fullTextLimit: string;
}

export const DEFAULT_HYBRID_CONFIG: HybridConfig = {
  vectorSearchOnly: false,
  fullTextWeight: "",
  semanticWeight: "",
  rrfK: "",
  fullTextLimit: "",
};

/** Converts panel state into request body fields, omitting anything left blank/default. */
export function buildHybridRequestFields(config: HybridConfig): {
  vector_search_only?: boolean;
  hybrid_settings?: HybridSettings;
} {
  const settings: HybridSettings = {};
  if (config.fullTextWeight !== "") settings.full_text_weight = parseFloat(config.fullTextWeight);
  if (config.semanticWeight !== "") settings.semantic_weight = parseFloat(config.semanticWeight);
  if (config.rrfK !== "") settings.rrf_k = parseFloat(config.rrfK);
  if (config.fullTextLimit !== "") settings.full_text_limit = parseInt(config.fullTextLimit);

  return {
    ...(config.vectorSearchOnly ? { vector_search_only: true } : {}),
    ...(Object.keys(settings).length > 0 ? { hybrid_settings: settings } : {}),
  };
}

export function HybridSettingsPanel({ value, onChange }: { value: HybridConfig; onChange: (v: HybridConfig) => void }) {
  const [expanded, setExpanded] = useState(false);

  function set<K extends keyof HybridConfig>(key: K, v: HybridConfig[K]) {
    onChange({ ...value, [key]: v });
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <button
          type="button"
          onClick={() => setExpanded(v => !v)}
          className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
        >
          <SlidersHorizontal className="w-3.5 h-3.5" />
          Hybrid search settings
        </button>
        <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer select-none">
          <input
            type="checkbox"
            checked={value.vectorSearchOnly}
            onChange={e => set("vectorSearchOnly", e.target.checked)}
            className="rounded"
          />
          Vector search only
        </label>
      </div>

      {expanded && !value.vectorSearchOnly && (
        <div className="grid grid-cols-4 gap-3 rounded-md border border-border bg-muted/20 p-3">
          <div className="space-y-1">
            <Label htmlFor="hybrid-full-text-weight" className="text-xs text-muted-foreground">Full-text weight</Label>
            <Input id="hybrid-full-text-weight" type="number" min={0} step={0.1} placeholder="0.3" value={value.fullTextWeight} onChange={e => set("fullTextWeight", e.target.value)} />
          </div>
          <div className="space-y-1">
            <Label htmlFor="hybrid-semantic-weight" className="text-xs text-muted-foreground">Semantic weight</Label>
            <Input id="hybrid-semantic-weight" type="number" min={0} step={0.1} placeholder="0.7" value={value.semanticWeight} onChange={e => set("semanticWeight", e.target.value)} />
          </div>
          <div className="space-y-1">
            <Label htmlFor="hybrid-rrf-k" className="text-xs text-muted-foreground">RRF k</Label>
            <Input id="hybrid-rrf-k" type="number" min={0.01} step={1} placeholder="60" value={value.rrfK} onChange={e => set("rrfK", e.target.value)} />
          </div>
          <div className="space-y-1">
            <Label htmlFor="hybrid-full-text-limit" className="text-xs text-muted-foreground">Full-text limit</Label>
            <Input id="hybrid-full-text-limit" type="number" min={1} max={1000} step={1} placeholder="200" value={value.fullTextLimit} onChange={e => set("fullTextLimit", e.target.value)} />
          </div>
        </div>
      )}
    </div>
  );
}
