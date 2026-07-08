"use client";

import { useEffect, useState } from "react";
import { Search, SlidersHorizontal } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api, Collection, MetadataField, QueryResultItem } from "@/lib/api";
import { ChunkCard } from "@/components/chunk-card";

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

export default function SearchPage() {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [slug, setSlug] = useState("");
  const [metadataFields, setMetadataFields] = useState<MetadataField[]>([]);
  const [query, setQuery] = useState("");
  const [topK, setTopK] = useState("5");
  const [filterText, setFilterText] = useState("");
  const [showFilters, setShowFilters] = useState(false);
  const [filterError, setFilterError] = useState<string | null>(null);
  const [results, setResults] = useState<QueryResultItem[] | null>(null);
  const [querying, setQuerying] = useState(false);
  const [showProperties, setShowProperties] = useState(false);

  useEffect(() => {
    api.collections.list().then(d => {
      const cols = d.collections ?? [];
      setCollections(cols);
      if (cols.length > 0) setSlug(cols[0].slug);
    });
  }, []);

  useEffect(() => {
    if (!slug) return;
    setMetadataFields([]);
    api.metadataFields.list(slug)
      .then(d => setMetadataFields(d.fields ?? []))
      .catch(() => {});
  }, [slug]);

  function parseFilters(): unknown | null {
    if (!filterText.trim()) return null;
    try {
      const parsed = JSON.parse(filterText);
      setFilterError(null);
      return parsed;
    } catch {
      setFilterError("Invalid JSON");
      return undefined;
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!slug) return;

    const filters = parseFilters();
    if (filters === undefined) return;

    setQuerying(true);
    setResults(null);
    try {
      const body: Parameters<typeof api.query.run>[1] = { query, top_k: parseInt(topK) };
      if (filters !== null) body.filters = filters;
      const data = await api.query.run(slug, body);
      setResults(data.results ?? []);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Query failed");
    } finally {
      setQuerying(false);
    }
  }

  function insertField(field: Pick<MetadataField, "name" | "type">) {
    const placeholder = field.type === "num" || field.type === "date"
      ? `{"${field.name}": {"$gte": 0}}`
      : field.type === "bool"
      ? `{"${field.name}": true}`
      : field.type === "arr"
      ? `{"${field.name}": {"$contains": "value"}}`
      : `{"${field.name}": "value"}`;

    setFilterText(prev => {
      if (!prev.trim()) return placeholder;
      try {
        const parsed = JSON.parse(prev);
        const snippet = field.type === "num" || field.type === "date"
          ? { "$gte": 0 }
          : field.type === "bool"
          ? true
          : field.type === "arr"
          ? { "$contains": "value" }
          : "value";
        return JSON.stringify({ ...parsed, [field.name]: snippet }, null, 2);
      } catch {
        return prev;
      }
    });
    setFilterError(null);
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="rounded-lg border bg-white p-6 space-y-4">
        <div className="flex gap-4">
          <div className="flex-1 space-y-1.5">
            <Label className="text-xs text-zinc-500">Collection</Label>
            <Select value={slug} onValueChange={v => v && setSlug(v)}>
              <SelectTrigger className="[&>span]:truncate">
                <SelectValue placeholder="Pick a collection" />
              </SelectTrigger>
              <SelectContent>
                {collections.map(c => (
                  <SelectItem key={c.slug} value={c.slug}>{c.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex-1 space-y-1.5">
            <Label className="text-xs text-zinc-500">Top K</Label>
            <Input
              type="number"
              min={1}
              max={100}
              value={topK}
              onChange={e => setTopK(e.target.value)}
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-zinc-500">Query</Label>
          <div className="flex gap-2">
            <Input
              required
              value={query}
              onChange={e => setQuery(e.target.value)}
              placeholder="What is machine learning?"
              className="flex-1"
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={() => setShowFilters(v => !v)}
              className={showFilters ? "border-primary text-primary" : ""}
              title="Toggle filters"
            >
              <SlidersHorizontal className="w-4 h-4" />
            </Button>
            <Button type="submit" disabled={querying || !slug} className="gap-2">
              <Search className="w-4 h-4" />
              {querying ? "Searching…" : "Search"}
            </Button>
          </div>
        </div>

        {showFilters && (
          <div className="space-y-2">
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
                className={`w-full rounded-md border px-3 py-2 text-sm font-mono resize-y min-h-[80px] focus:outline-none focus:ring-2 focus:ring-primary/30 ${filterError ? "border-red-400" : "border-zinc-200"}`}
                placeholder={'{"author": "Alice", "score": {"$gte": 4}}'}
                value={filterText}
                onChange={e => { setFilterText(e.target.value); setFilterError(null); }}
              />
              {filterError && <p className="text-xs text-red-500">{filterError}</p>}
            </div>
          </div>
        )}
      </form>

      {results !== null && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <p className="text-sm text-zinc-500">{results.length} result{results.length !== 1 ? "s" : ""}</p>
            {results.some(r => r.extra_json) && (
              <label className="flex items-center gap-2 text-xs text-zinc-500 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={showProperties}
                  onChange={e => setShowProperties(e.target.checked)}
                  className="rounded"
                />
                Show file properties
              </label>
            )}
          </div>
          {results.length === 0 ? (
            <div className="rounded-lg border bg-white px-6 py-10 text-center text-zinc-400 text-sm">
              No results found.
            </div>
          ) : results.map((r, i) => (
            <ChunkCard
              key={i}
              chunkIndex={r.chunk_index}
              source={r.source}
              fileUri={r.file_uri}
              similarity={r.similarity}
              chunkHeader={r.chunk_header}
              chunkText={r.chunk_text}
              metadata={r.metadata}
              extraJSON={showProperties ? r.extra_json : null}
            />
          ))}
        </div>
      )}
    </div>
  );
}
