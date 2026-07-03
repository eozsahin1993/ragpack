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
  const [ingesting, setIngesting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIngesting(true);
    try {
      await api.ingest.uri(slug, { file_uri: uri, mime_type: "" });
      setUri("");
      onComplete();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ingest failed");
    } finally {
      setIngesting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex gap-2">
      <Input
        required
        value={uri}
        onChange={e => setUri(e.target.value)}
        placeholder="https://… or s3://bucket/key"
        className="flex-1"
      />
      <Button type="submit" disabled={ingesting}>
        {ingesting ? "Ingesting…" : "Ingest"}
      </Button>
    </form>
  );
}
