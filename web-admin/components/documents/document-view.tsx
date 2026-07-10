"use client";

import { useEffect, useRef, useState } from "react";
import { Pencil, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { api, Chunk, Document, MetadataField } from "@/lib/api";
import { Pagination } from "@/components/pagination";
import { ChunkCard } from "@/components/chunk-card";
import { DocumentDetails } from "./document-details";
import { useBreadcrumbLabel } from "@/components/breadcrumb-context";
import { friendlyUri } from "@/lib/utils";

const PAGE_SIZE = 20;

function docLabel(doc: Document) {
  return doc.name ?? friendlyUri(doc.file_uri);
}

interface DocumentViewProps {
  // Pass the collection slug when reached via /collections/:slug/documents/:id;
  // pass null when reached via the slug-less /documents/:id route — the slug
  // is then resolved once (document -> collection_id -> collection.slug).
  slug: string | null;
  id: string;
  onDeleted: () => void;
}

export function DocumentView({ slug: slugProp, id, onDeleted }: DocumentViewProps) {
  const [slug, setSlug] = useState<string | null>(slugProp);
  const [doc, setDoc] = useState<Document | null>(null);
  const [chunks, setChunks] = useState<Chunk[]>([]);
  const [metadataFields, setMetadataFields] = useState<MetadataField[]>([]);
  const [currentMetadata, setCurrentMetadata] = useState<Record<string, unknown>>({});
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
      onDeleted();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Delete failed");
      setDeleting(false);
    }
  }

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const d = await api.documents.get(slugProp, id);
        let resolvedSlug = slugProp;
        if (!resolvedSlug) {
          const col = await api.collections.getById(d.collection_id);
          resolvedSlug = col.slug;
        }
        if (cancelled) return;
        setSlug(resolvedSlug);

        const [c, mf, md] = await Promise.all([
          api.documents.chunks(resolvedSlug, id),
          api.metadataFields.list(resolvedSlug),
          api.documents.metadata(resolvedSlug, id),
        ]);
        if (cancelled) return;

        setDoc(d);
        setChunks(c.chunks ?? []);
        setMetadataFields(mf.fields ?? []);
        setCurrentMetadata(md.metadata ?? {});
        const label = docLabel(d);
        setBreadcrumbLabel(id, label.length > 30 ? label.slice(0, 30) + "…" : label);
      } catch (e) {
        if (!cancelled) toast.error(e instanceof Error ? e.message : "Failed to load");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => { cancelled = true; };
  }, [slugProp, id]); // eslint-disable-line react-hooks/exhaustive-deps

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
              <button onClick={() => setEditingName(false)} className="text-xs text-muted-foreground hover:underline">Cancel</button>
            </div>
          ) : (
            <div className="flex items-center gap-2 min-w-0">
              <h1 className="text-xl font-semibold truncate">{doc ? docLabel(doc) : "Document"}</h1>
              {doc && (
                <button onClick={startEdit} className="text-muted-foreground/50 hover:text-primary shrink-0">
                  <Pencil className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          )}
          {doc && <p className="text-xs text-muted-foreground font-mono mt-0.5 truncate">{friendlyUri(doc.file_uri)}</p>}
        </div>
        {doc && (
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive hover:bg-destructive/10 shrink-0"
            onClick={handleDelete}
            disabled={deleting}
          >
            <Trash2 className="w-4 h-4 mr-1.5" />
            {deleting ? "Deleting…" : "Delete"}
          </Button>
        )}
      </div>

      {doc && (
        <DocumentDetails
          doc={doc}
          slug={slug}
          metadataFields={metadataFields}
          currentMetadata={currentMetadata}
          onUpdate={setDoc}
          onMetadataUpdate={setCurrentMetadata}
        />
      )}

      {/* Chunks */}
      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : chunks.length === 0 ? (
        <p className="text-sm text-muted-foreground">No chunks found for this document.</p>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">{chunks.length} chunks</p>
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
