"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";

interface UriIngestProps {
  slug: string;
  onComplete: () => void;
}

export function UriIngest({ slug, onComplete }: UriIngestProps) {
  const [uri, setUri] = useState("");
  const [extraJSON, setExtraJSON] = useState("");
  const [jsonError, setJsonError] = useState(false);
  const [ingesting, setIngesting] = useState(false);

  function handleExtraJSONChange(value: string) {
    setExtraJSON(value);
    if (value.trim() === "") {
      setJsonError(false);
    } else {
      try {
        JSON.parse(value);
        setJsonError(false);
      } catch {
        setJsonError(true);
      }
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (jsonError) return;
    setIngesting(true);
    try {
      await api.ingest.uri(slug, {
        file_uri: uri,
        mime_type: "",
        extra_json: extraJSON.trim() || undefined,
      });
      setUri("");
      setExtraJSON("");
      onComplete();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ingest failed");
    } finally {
      setIngesting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <div className="flex gap-2">
        <Input
          required
          value={uri}
          onChange={e => setUri(e.target.value)}
          placeholder="https://… or s3://bucket/key"
          className="flex-1"
        />
        <Button type="submit" disabled={ingesting || jsonError}>
          {ingesting ? "Ingesting…" : "Ingest"}
        </Button>
      </div>
      <div>
        <textarea
          value={extraJSON}
          onChange={e => handleExtraJSONChange(e.target.value)}
          placeholder='Metadata JSON (optional) — e.g. {"author": "Alice"}'
          rows={2}
          className={`w-full rounded-md border px-3 py-2 text-xs font-mono resize-none bg-white placeholder:text-zinc-400 focus:outline-none focus:ring-1 ${
            jsonError
              ? "border-red-300 focus:ring-red-300"
              : "border-zinc-200 focus:ring-zinc-300"
          }`}
        />
        {jsonError && (
          <p className="text-xs text-red-500 mt-0.5">Must be valid JSON</p>
        )}
      </div>
    </form>
  );
}
