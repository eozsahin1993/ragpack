"use client";

import { useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Label } from "@/components/ui/label";
import { MetadataField } from "@/lib/api";

const TYPE_LABELS: Record<MetadataField["type"], string> = {
  str: "str",
  num: "num",
  bool: "bool",
  date: "date",
  arr: "arr[]",
};

const BUILTIN_FIELDS: { name: string; type: MetadataField["type"] }[] = [
  { name: "mime_type", type: "str" },
  { name: "source_name", type: "str" },
  { name: "external_id", type: "str" },
  { name: "file_uri", type: "str" },
  { name: "created_at", type: "date" },
  { name: "updated_at", type: "date" },
];

/** Parses filter panel text into a filters object: null if empty, undefined if invalid JSON. */
export function parseFilterText(text: string): unknown | null | undefined {
  if (!text.trim()) return null;
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

interface FilterPanelProps {
  value: string;
  onChange: (v: string) => void;
  error?: string | null;
  metadataFields: MetadataField[];
}

export function FilterPanel({ value, onChange, error, metadataFields }: FilterPanelProps) {
  const [expanded, setExpanded] = useState(false);

  function insertField(field: Pick<MetadataField, "name" | "type">) {
    const placeholder = field.type === "num" || field.type === "date"
      ? `{"${field.name}": {"$gte": 0}}`
      : field.type === "bool"
      ? `{"${field.name}": true}`
      : field.type === "arr"
      ? `{"${field.name}": {"$contains": "value"}}`
      : `{"${field.name}": "value"}`;

    if (!value.trim()) {
      onChange(placeholder);
      return;
    }
    try {
      const parsed = JSON.parse(value);
      const snippet = field.type === "num" || field.type === "date"
        ? { "$gte": 0 }
        : field.type === "bool"
        ? true
        : field.type === "arr"
        ? { "$contains": "value" }
        : "value";
      onChange(JSON.stringify({ ...parsed, [field.name]: snippet }, null, 2));
    } catch {
      // current text isn't valid JSON to merge into — leave it for the user to fix
    }
  }

  return (
    <div className="space-y-2">
      <button
        type="button"
        onClick={() => setExpanded(v => !v)}
        className={`flex items-center gap-1.5 text-xs transition-colors ${expanded ? "text-primary" : "text-muted-foreground hover:text-foreground"}`}
      >
        <SlidersHorizontal className="w-3.5 h-3.5" />
        Document Filters
      </button>

      {expanded && (
        <div className="space-y-2 rounded-md border border-border bg-muted/20 p-3">
          <div className="space-y-1.5">
            <p className="text-xs text-zinc-400">Filterable fields — click to insert</p>
            <div className="flex flex-wrap gap-1.5">
              {BUILTIN_FIELDS.map(f => (
                <button
                  key={f.name}
                  type="button"
                  onClick={() => insertField(f)}
                  className="inline-flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-2.5 py-1 text-xs hover:border-zinc-400 hover:bg-zinc-50 transition-colors"
                >
                  <span className="font-medium text-zinc-500">{f.name}</span>
                  <span className="text-zinc-300">{TYPE_LABELS[f.type]}</span>
                </button>
              ))}
              {metadataFields.map(f => (
                <button
                  key={f.id}
                  type="button"
                  onClick={() => insertField(f)}
                  className="inline-flex items-center gap-1.5 rounded-full border border-zinc-300 bg-zinc-50 px-2.5 py-1 text-xs hover:border-zinc-500 hover:bg-zinc-100 transition-colors"
                >
                  <span className="font-medium text-zinc-700">{f.name}</span>
                  <span className="text-zinc-400">{TYPE_LABELS[f.type]}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs text-zinc-500">Filter (MongoDB-style JSON)</Label>
            <textarea
              className={`w-full rounded-md border px-3 py-2 text-sm font-mono resize-y min-h-[80px] focus:outline-none focus:ring-2 focus:ring-primary/30 ${error ? "border-red-400" : "border-zinc-200"}`}
              placeholder={'{"author": "Alice", "score": {"$gte": 4}}'}
              value={value}
              onChange={e => onChange(e.target.value)}
            />
            {error && <p className="text-xs text-red-500">{error}</p>}
          </div>
        </div>
      )}
    </div>
  );
}
