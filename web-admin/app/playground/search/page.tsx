"use client";

import { useEffect, useState } from "react";
import { Search } from "lucide-react";
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
import { DEFAULT_HYBRID_CONFIG, HybridConfig, HybridSettingsPanel, buildHybridRequestFields } from "@/components/hybrid-settings-panel";
import { FilterPanel, parseFilterText } from "@/components/filter-panel";

export default function SearchPage() {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [slug, setSlug] = useState("");
  const [metadataFields, setMetadataFields] = useState<MetadataField[]>([]);
  const [query, setQuery] = useState("");
  const [topK, setTopK] = useState("5");
  const [filterText, setFilterText] = useState("");
  const [filterError, setFilterError] = useState<string | null>(null);
  const [results, setResults] = useState<QueryResultItem[] | null>(null);
  const [querying, setQuerying] = useState(false);
  const [showProperties, setShowProperties] = useState(false);
  const [hybridConfig, setHybridConfig] = useState<HybridConfig>(DEFAULT_HYBRID_CONFIG);

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

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!slug) return;

    const filters = parseFilterText(filterText);
    if (filters === undefined) {
      setFilterError("Invalid JSON");
      return;
    }
    setFilterError(null);

    setQuerying(true);
    setResults(null);
    try {
      const body: Parameters<typeof api.query.run>[1] = {
        query,
        top_k: parseInt(topK),
        ...buildHybridRequestFields(hybridConfig),
      };
      if (filters !== null) body.filters = filters;
      const data = await api.query.run(slug, body);
      setResults(data.results ?? []);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Query failed");
    } finally {
      setQuerying(false);
    }
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
            <Button type="submit" disabled={querying || !slug} className="gap-2">
              <Search className="w-4 h-4" />
              {querying ? "Searching…" : "Search"}
            </Button>
          </div>
        </div>

        <HybridSettingsPanel value={hybridConfig} onChange={setHybridConfig} />
        <FilterPanel value={filterText} onChange={setFilterText} error={filterError} metadataFields={metadataFields} />
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
              similarity={r.vector_similarity}
              keywordScore={r.keyword_bm25_score}
              rrfScoreNormalized={r.rrf_score_normalized}
              rrfScore={r.rrf_score}
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
