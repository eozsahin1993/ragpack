"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Pencil, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { api, Chunk, Document } from "@/lib/api"; // Document used by useState type inference
import { Pagination } from "@/components/pagination";
import { ChunkCard } from "@/components/chunk-card";
import { DocumentDetails } from "./_components/document-details";
import { useBreadcrumbLabel } from "@/components/breadcrumb-context";

const PAGE_SIZE = 20;


function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

function docLabel(doc: Document) {
  return doc.name ?? friendlyUri(doc.file_uri);
}

export default function ChunksPage() {
  const { slug, id } = useParams<{ slug: string; id: string }>();
  const router = useRouter();

  const [doc, setDoc] = useState<Document | null>(null);
  const [chunks, setChunks] = useState<Chunk[]>([]);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);
  const [editingName, setEditingName] = useState(false);
  const [nameInput, setNameInput] = useState("");
  const nameRef = useRef<HTMLInputElement>(null);
  const setBreadcrumbLabel = useBreadcrumbLabel();

  function startEdit() {
    if (!doc) return;
    setNameInput(doc.name ?? docLabel(doc));
    setEditingName(true);
    setTimeout(() => nameRef.current?.select(), 0);
  }

  async function handleSaveName() {
    if (!doc || !nameInput.trim()) return;
    try {
      const updated = await api.documents.update(slug, id, { name: nameInput.trim() });
      setDoc(updated);
      setEditingName(false);
      const label = docLabel(updated);
      setBreadcrumbLabel(id, label.length > 30 ? label.slice(0, 30) + "…" : label);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to update name");
    }
  }

  async function handleDelete() {
    if (!doc) return;
    if (!confirm(`Delete "${docLabel(doc)}"? This removes all indexed chunks.`)) return;
    setDeleting(true);
    try {
      await api.documents.delete(slug, id);
      router.push(`/collections/${slug}`);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Delete failed");
      setDeleting(false);
    }
  }

  useEffect(() => {
    Promise.all([
      api.documents.get(slug, id),
      api.documents.chunks(slug, id),
    ])
      .then(([d, c]) => {
        setDoc(d);
        setChunks(c.chunks ?? []);
        const label = docLabel(d);
        setBreadcrumbLabel(id, label.length > 30 ? label.slice(0, 30) + "…" : label);
      })
      .catch(e => toast.error(e instanceof Error ? e.message : "Failed to load"))
      .finally(() => setLoading(false));
  }, [slug, id]);

  const totalPages = Math.ceil(chunks.length / PAGE_SIZE);
  const visibleChunks = chunks.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          {editingName ? (
            <div className="flex items-center gap-2">
              <input
                ref={nameRef}
                value={nameInput}
                onChange={e => setNameInput(e.target.value)}
                onKeyDown={e => { if (e.key === "Enter") handleSaveName(); if (e.key === "Escape") setEditingName(false); }}
                className="text-xl font-semibold border-b border-primary outline-none bg-transparent w-80"
                autoFocus
              />
              <button onClick={handleSaveName} className="text-xs text-primary hover:underline">Save</button>
              <button onClick={() => setEditingName(false)} className="text-xs text-zinc-400 hover:underline">Cancel</button>
            </div>
          ) : (
            <div className="flex items-center gap-2 min-w-0">
              <h1 className="text-xl font-semibold truncate">{doc ? docLabel(doc) : "Chunks"}</h1>
              {doc && (
                <button onClick={startEdit} className="text-zinc-300 hover:text-zinc-500 shrink-0">
                  <Pencil className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          )}
          {doc && <p className="text-xs text-zinc-400 font-mono mt-0.5 truncate">{friendlyUri(doc.file_uri)}</p>}
        </div>
        {doc && (
          <Button
            variant="ghost"
            size="sm"
            className="text-red-500 hover:text-red-600 hover:bg-red-50 shrink-0"
            onClick={handleDelete}
            disabled={deleting}
          >
            <Trash2 className="w-4 h-4 mr-1.5" />
            {deleting ? "Deleting…" : "Delete"}
          </Button>
        )}
      </div>

      {doc && <DocumentDetails doc={doc} />}

      {/* Chunks */}
      {loading ? (
        <p className="text-sm text-zinc-400">Loading…</p>
      ) : chunks.length === 0 ? (
        <p className="text-sm text-zinc-400">No chunks found for this document.</p>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-zinc-500">{chunks.length} chunks</p>
            <Pagination page={page} totalPages={totalPages} total={chunks.length} pageSize={PAGE_SIZE} onPageChange={setPage} />
          </div>

          <div className="space-y-3">
            {visibleChunks.map(ch => (
              <ChunkCard
                key={ch.id}
                chunkIndex={ch.chunk_index}
                chunkHeader={ch.chunk_header}
                chunkText={ch.chunk_text}
                chunkHash={ch.chunk_hash}
              />
            ))}
          </div>

          <Pagination page={page} totalPages={totalPages} total={chunks.length} pageSize={PAGE_SIZE} onPageChange={setPage} />
        </div>
      )}
    </div>
  );
}
