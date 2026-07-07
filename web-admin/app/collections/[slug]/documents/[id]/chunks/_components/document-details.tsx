"use client";

import { useState } from "react";
import { Pencil } from "lucide-react";
import { toast } from "sonner";
import { api, Document } from "@/lib/api";
import { timeAgo } from "@/lib/utils";

const statusColors: Record<string, string> = {
  complete:  "bg-emerald-50 text-emerald-700 border-emerald-200",
  ingesting: "bg-amber-50 text-amber-700 border-amber-200",
  failed:    "bg-red-50 text-red-700 border-red-200",
};

interface DocumentDetailsProps {
  doc: Document;
  slug: string;
  onUpdate: (updated: Document) => void;
}

export function DocumentDetails({ doc, slug, onUpdate }: DocumentDetailsProps) {
  const [editing, setEditing] = useState(false);
  const [input, setInput] = useState("");
  const [jsonError, setJsonError] = useState(false);
  const [saving, setSaving] = useState(false);

  function startEdit() {
    setInput(doc.extra_json ? formatJSON(doc.extra_json) : "");
    setJsonError(false);
    setEditing(true);
  }

  function handleChange(value: string) {
    setInput(value);
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

  async function handleSave() {
    if (jsonError || saving) return;
    setSaving(true);
    try {
      const updated = await api.documents.update(slug, doc.id, { extra_json: input.trim() || undefined });
      onUpdate(updated);
      setEditing(false);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to update");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="rounded-lg border bg-white p-4 grid grid-cols-2 gap-x-8 gap-y-2 text-sm sm:grid-cols-4">
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Type</p>
        <p className="text-zinc-700 font-mono text-xs">{doc.mime_type}</p>
      </div>
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Status</p>
        <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[doc.status] ?? ""}`}>
          {doc.status}
        </span>
      </div>
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Chunks</p>
        <p className="text-zinc-700">{doc.chunk_count}</p>
      </div>
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Ingested</p>
        <p className="text-zinc-700" title={new Date(doc.created_at).toLocaleString()}>{timeAgo(doc.created_at)}</p>
      </div>

      <div className="col-span-2 sm:col-span-4">
        <div className="flex items-center justify-between mb-1">
          <p className="text-xs text-zinc-400">File properties</p>
          {!editing && (
            <button onClick={startEdit} className="text-zinc-300 hover:text-zinc-500 transition-colors">
              <Pencil className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        {editing ? (
          <div className="space-y-1.5">
            <textarea
              value={input}
              onChange={e => handleChange(e.target.value)}
              placeholder='{"author": "Alice", "department": "eng"}'
              rows={4}
              autoFocus
              className={`w-full rounded-md border px-3 py-2 text-xs font-mono resize-y bg-white placeholder:text-zinc-300 focus:outline-none focus:ring-1 ${
                jsonError ? "border-red-300 focus:ring-red-300" : "border-zinc-200 focus:ring-zinc-300"
              }`}
            />
            {jsonError && <p className="text-xs text-red-500">Must be valid JSON</p>}
            <div className="flex gap-3">
              <button
                onClick={handleSave}
                disabled={jsonError || saving}
                className="text-xs text-primary hover:underline disabled:opacity-40"
              >
                {saving ? "Saving…" : "Save"}
              </button>
              <button onClick={() => setEditing(false)} className="text-xs text-zinc-400 hover:underline">
                Cancel
              </button>
            </div>
          </div>
        ) : doc.extra_json ? (
          <pre className="text-xs font-mono bg-zinc-50 border border-zinc-100 rounded p-2 overflow-x-auto text-zinc-700 whitespace-pre-wrap break-all">
            {formatJSON(doc.extra_json)}
          </pre>
        ) : (
          <p className="text-xs text-zinc-300 italic">None</p>
        )}
      </div>

      {doc.error && (
        <div className="col-span-2 sm:col-span-4">
          <p className="text-xs text-red-400">{doc.error}</p>
        </div>
      )}
    </div>
  );
}

function formatJSON(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}
