"use client";

import { useEffect, useState } from "react";
import { Search } from "lucide-react";
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
import { api, Collection, QueryResultItem } from "@/lib/api";
import { ChunkCard } from "@/components/chunk-card";

export default function SearchPage() {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [slug, setSlug] = useState("");
  const [query, setQuery] = useState("");
  const [topK, setTopK] = useState("5");
  const [results, setResults] = useState<QueryResultItem[] | null>(null);
  const [querying, setQuerying] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    api.collections.list().then(d => {
      const cols = d.collections ?? [];
      setCollections(cols);
      if (cols.length > 0) setSlug(cols[0].slug);
    });
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!slug) return;
    setQuerying(true);
    setError("");
    setResults(null);
    try {
      const data = await api.query.run(slug, { query, top_k: parseInt(topK) });
      setResults(data.results ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Query failed");
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

        {error && <p className="text-red-500 text-sm">{error}</p>}
      </form>

      {results !== null && (
        <div className="space-y-3">
          <p className="text-sm text-zinc-500">{results.length} result{results.length !== 1 ? "s" : ""}</p>
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
            />
          ))}
        </div>
      )}
    </div>
  );
}
