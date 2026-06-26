"use client";

import { useEffect, useState } from "react";
import { Check, Copy, Sparkles } from "lucide-react";
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
import { api, Collection, Prompt, RagResponse } from "@/lib/api";

export default function RagPage() {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [llmModels, setLlmModels] = useState<string[]>([]);
  const [slug, setSlug] = useState("");
  const [query, setQuery] = useState("");
  const [topK, setTopK] = useState("5");
  const [promptSlug, setPromptSlug] = useState("");
  const [model, setModel] = useState("");
  const [ragResult, setRagResult] = useState<RagResponse | null>(null);
  const [querying, setQuerying] = useState(false);
  const [error, setError] = useState("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    api.collections.list().then(d => {
      const cols = d.collections ?? [];
      setCollections(cols);
      if (cols.length > 0) setSlug(cols[0].slug);
    });
    api.prompts.list().then(d => {
      const all = [...(d.system ?? []), ...(d.user ?? [])];
      setPrompts(all);
      if (all.length > 0) setPromptSlug(all[0].slug);
    });
    api.llms.list().then(d => {
      setLlmModels(d.models ?? []);
      setModel(d.default ?? "");
    }).catch(() => {
      // no LLM configured — playground will show a warning
    });
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!slug || !promptSlug) return;
    setQuerying(true);
    setError("");
    setRagResult(null);
    try {
      const data = await api.query.rag(slug, {
        query,
        top_k: parseInt(topK),
        prompt_slug: promptSlug,
        ...(model ? { model } : {}),
      });
      setRagResult(data);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "RAG failed");
    } finally {
      setQuerying(false);
    }
  }

  function copyPrompt() {
    if (!ragResult) return;
    navigator.clipboard.writeText(ragResult.formatted_prompt);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="rounded-lg border bg-white p-6 space-y-4">
        <div className="flex gap-4 flex-wrap">
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
          <div className="w-56 space-y-1.5">
            <Label className="text-xs text-zinc-500">Prompt</Label>
            {prompts.length === 0 ? (
              <p className="text-xs text-zinc-400 pt-2">No prompts yet — create one in Prompts.</p>
            ) : (
              <Select value={promptSlug} onValueChange={v => v && setPromptSlug(v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Pick a prompt" />
                </SelectTrigger>
                <SelectContent>
                  {prompts.map(p => (
                    <SelectItem key={p.slug} value={p.slug}>{p.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>
          <div className="w-56 space-y-1.5">
            <Label className="text-xs text-zinc-500">LLM Model</Label>
            {llmModels.length === 0 ? (
              <p className="text-xs text-zinc-400 pt-2">No LLM configured — set one in .env.</p>
            ) : (
              <Select value={model} onValueChange={v => v && setModel(v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Server default" />
                </SelectTrigger>
                <SelectContent>
                  {llmModels.map(m => (
                    <SelectItem key={m} value={m}>{m}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
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
            <Button type="submit" disabled={querying || !slug || !promptSlug || llmModels.length === 0} className="gap-2">
              <Sparkles className="w-4 h-4" />
              {querying ? "Running…" : "Run RAG"}
            </Button>
          </div>
        </div>

        {error && <p className="text-red-500 text-sm">{error}</p>}
      </form>

      {ragResult !== null && (
        <div className="space-y-4">
          {ragResult.answer && (
            <div className="rounded-lg border bg-white p-5 space-y-3">
              <h2 className="text-sm font-semibold">Answer</h2>
              <p className="text-sm text-zinc-700 leading-relaxed whitespace-pre-wrap">{ragResult.answer}</p>
            </div>
          )}

          <div className="rounded-lg border bg-white p-5 space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold">Formatted Prompt</h2>
              <Button type="button" variant="ghost" size="sm" onClick={copyPrompt} className="gap-1.5 text-xs h-7">
                {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
            <pre className="text-xs text-zinc-700 bg-zinc-50 rounded-md p-4 overflow-auto whitespace-pre-wrap leading-relaxed font-mono max-h-96">
              {ragResult.formatted_prompt}
            </pre>
          </div>

          <div className="space-y-3">
            <p className="text-sm text-zinc-500">{ragResult.chunks.length} source{ragResult.chunks.length !== 1 ? "s" : ""}</p>
            {ragResult.chunks.map((c, i) => (
              <div key={i} className="rounded-lg border bg-white p-5 space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{c.source}</span>
                    <span className="text-xs text-zinc-400">chunk {c.chunk_index}</span>
                  </div>
                  <div className="text-right">
                    <span className="text-lg font-semibold text-emerald-600">{c.similarity.toFixed(1)}%</span>
                    <p className="text-xs text-zinc-400">similarity</p>
                  </div>
                </div>
                {c.chunk_text && (
                  <p className="text-sm text-zinc-600 leading-relaxed border-t pt-3 mt-2">{c.chunk_text}</p>
                )}
                <p className="text-xs text-zinc-400 font-mono">{c.file_uri}</p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
