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

export default function PlaygroundPage() {
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

  async function handleQuery(e: React.FormEvent) {
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
    <div className="space-y-6 max-w-3xl">
      <div>
        <h1 className="text-xl font-semibold">Playground</h1>
        <p className="text-sm text-zinc-500 mt-0.5">Run semantic queries against your collections</p>
      </div>

      <form onSubmit={handleQuery} className="rounded-lg border bg-white p-6 space-y-4">
        <div className="flex gap-4">
          <div className="w-56 space-y-1.5">
            <Label className="text-xs text-zinc-500">Collection</Label>
            <Select value={slug} onValueChange={v => v && setSlug(v)}>
              <SelectTrigger>
                <SelectValue placeholder="Pick a collection" />
              </SelectTrigger>
              <SelectContent>
                {collections.map(c => (
                  <SelectItem key={c.slug} value={c.slug}>{c.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-24 space-y-1.5">
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
            <div key={i} className="rounded-lg border bg-white p-5 space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">{r.source}</span>
                  <span className="text-xs text-zinc-400">chunk {r.chunk_index}</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="text-right">
                    <span className="text-lg font-semibold text-emerald-600">{r.similarity.toFixed(1)}%</span>
                    <p className="text-xs text-zinc-400">similarity</p>
                  </div>
                </div>
              </div>
              {r.chunk_text && (
                <p className="text-sm text-zinc-600 leading-relaxed border-t pt-3 mt-2">{r.chunk_text}</p>
              )}
              <p className="text-xs text-zinc-400 font-mono">{r.file_uri}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
