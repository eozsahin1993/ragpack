"use client";

import { useState } from "react";
import { Pencil } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { api, Document, MetadataField } from "@/lib/api";
import { MetaFieldInput } from "@/components/documents/meta-field-input";
import { timeAgo } from "@/lib/utils";

const statusColors: Record<string, string> = {
  complete:  "bg-emerald-50 text-emerald-700 border-emerald-200",
  ingesting: "bg-amber-50 text-amber-700 border-amber-200",
  failed:    "bg-red-50 text-red-700 border-red-200",
};

interface DocumentDetailsProps {
  doc: Document;
  slug: string | null;
  metadataFields: MetadataField[];
  currentMetadata: Record<string, unknown>;
  onUpdate: (updated: Document) => void;
  onMetadataUpdate: (metadata: Record<string, unknown>) => void;
}

export function DocumentDetails({ doc, slug, metadataFields, currentMetadata, onUpdate, onMetadataUpdate }: DocumentDetailsProps) {
  const [editing, setEditing] = useState(false);
  const [input, setInput] = useState("");
  const [jsonError, setJsonError] = useState(false);
  const [saving, setSaving] = useState(false);

  const [editingMeta, setEditingMeta] = useState(false);
  const [metaInputs, setMetaInputs] = useState<Record<string, string>>({});
  const [savingMeta, setSavingMeta] = useState(false);

  function startEdit() {
    setInput(doc.extra_json ? formatJSON(doc.extra_json) : "");
    setJsonError(false);
    setEditing(true);
  }

  function handleChange(value: string) {
    setInput(value);
    if (value.trim() === "") { setJsonError(false); return; }
    try { JSON.parse(value); setJsonError(false); }
    catch { setJsonError(true); }
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

  function startEditMeta() {
    setMetaInputs(Object.fromEntries(metadataFields.map(f => {
      const v = currentMetadata[f.name];
      if (v === undefined || v === null) return [f.name, ""];
      if (Array.isArray(v)) return [f.name, v.join(", ")];
      return [f.name, String(v)];
    })));
    setEditingMeta(true);
  }

  async function handleSaveMeta() {
    if (savingMeta) return;
    setSavingMeta(true);
    try {
      const metadata: Record<string, unknown> = {};
      for (const field of metadataFields) {
        const raw = metaInputs[field.name]?.trim();
        if (!raw) continue;
        metadata[field.name] = coerceForType(raw, field.type);
      }
      if (Object.keys(metadata).length === 0) { setEditingMeta(false); return; }
      const updated = await api.documents.update(slug, doc.id, { metadata });
      onUpdate(updated);
      const fresh = await api.documents.metadata(slug, doc.id);
      onMetadataUpdate(fresh.metadata);
      setEditingMeta(false);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to update properties");
    } finally {
      setSavingMeta(false);
    }
  }

  return (
    <div className="space-y-3">

      {/* Stats */}
      <Card>
        <CardContent className="grid grid-cols-2 sm:grid-cols-4 gap-x-8 gap-y-2 pt-4 pb-4">
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
            <p className="text-zinc-700 text-sm">{doc.chunk_count}</p>
          </div>
          <div>
            <p className="text-xs text-zinc-400 mb-0.5">Ingested</p>
            <p className="text-zinc-700 text-sm" title={new Date(doc.created_at).toLocaleString()}>{timeAgo(doc.created_at)}</p>
          </div>
          {doc.error && (
            <div className="col-span-2 sm:col-span-4">
              <p className="text-xs text-red-400">{doc.error}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Properties */}
      {metadataFields.length > 0 && (
        <Card>
          <CardHeader className="pb-2 pt-4 px-4">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium">Properties</CardTitle>
              {!editingMeta && (
                <button onClick={startEditMeta} className="text-zinc-300 hover:text-zinc-500 transition-colors">
                  <Pencil className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            {editingMeta ? (
              <div className="space-y-3">
                {metadataFields.map(field => (
                  <div key={field.id} className="flex items-center gap-3">
                    <div className="w-36 shrink-0">
                      <Label className="text-xs font-medium text-zinc-700">{field.name}</Label>
                      <p className="text-[10px] text-zinc-400 mt-0.5">{field.type}</p>
                    </div>
                    <div className="flex-1">
                      <MetaFieldInput
                        field={field}
                        value={metaInputs[field.name] ?? ""}
                        onChange={v => setMetaInputs(prev => ({ ...prev, [field.name]: v }))}
                      />
                    </div>
                  </div>
                ))}
                <div className="flex gap-2 pt-1">
                  <Button size="sm" onClick={handleSaveMeta} disabled={savingMeta}>
                    {savingMeta ? "Saving…" : "Save"}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setEditingMeta(false)}>Cancel</Button>
                </div>
              </div>
            ) : (
              <div className="divide-y divide-zinc-50">
                {metadataFields.map(f => {
                  const v = currentMetadata[f.name];
                  return (
                    <div key={f.id} className="flex items-center justify-between py-1.5 first:pt-0 last:pb-0">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-medium text-zinc-700">{f.name}</span>
                        <Badge variant="outline" className="text-[10px] font-normal text-zinc-400 px-1.5 py-0">
                          {f.type}
                        </Badge>
                      </div>
                      <span className="text-xs font-mono text-zinc-600">
                        {v == null
                          ? <span className="text-zinc-300">—</span>
                          : Array.isArray(v) ? v.join(", ") : String(v)}
                      </span>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* JSON extras */}
      <Card>
        <CardHeader className="pb-2 pt-4 px-4">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">JSON extras</CardTitle>
            {!editing && (
              <button onClick={startEdit} className="text-zinc-300 hover:text-zinc-500 transition-colors">
                <Pencil className="w-3.5 h-3.5" />
              </button>
            )}
          </div>
        </CardHeader>
        <CardContent className="px-4 pb-4">
          {editing ? (
            <div className="space-y-2">
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
              <div className="flex gap-2">
                <Button size="sm" onClick={handleSave} disabled={jsonError || saving}>
                  {saving ? "Saving…" : "Save"}
                </Button>
                <Button size="sm" variant="ghost" onClick={() => setEditing(false)}>Cancel</Button>
              </div>
            </div>
          ) : doc.extra_json ? (
            <pre className="text-xs font-mono bg-zinc-50 border border-zinc-100 rounded p-2 overflow-x-auto text-zinc-700 whitespace-pre-wrap break-all">
              {formatJSON(doc.extra_json)}
            </pre>
          ) : (
            <p className="text-xs text-zinc-300 italic">None</p>
          )}
        </CardContent>
      </Card>

    </div>
  );
}

function formatJSON(raw: string): string {
  try { return JSON.stringify(JSON.parse(raw), null, 2); }
  catch { return raw; }
}

function coerceForType(raw: string, type: MetadataField["type"]): unknown {
  switch (type) {
    case "num":  return parseFloat(raw);
    case "bool": return raw === "true";
    case "arr":  return raw.split(",").map(s => s.trim()).filter(Boolean);
    default:     return raw;
  }
}
