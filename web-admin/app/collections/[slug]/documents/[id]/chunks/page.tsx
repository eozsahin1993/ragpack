"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ChevronLeft, ChevronRight, Trash2, Hash } from "lucide-react";
import { Button } from "@/components/ui/button";
import { api, Chunk, Document } from "@/lib/api";

const PAGE_SIZE = 20;

const statusColors: Record<string, string> = {
  complete:  "bg-emerald-50 text-emerald-700 border-emerald-200",
  ingesting: "bg-amber-50 text-amber-700 border-amber-200",
  failed:    "bg-red-50 text-red-700 border-red-200",
};

function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

export default function ChunksPage() {
  const { slug, id } = useParams<{ slug: string; id: string }>();
  const router = useRouter();

  const [doc, setDoc] = useState<Document | null>(null);
  const [chunks, setChunks] = useState<Chunk[]>([]);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");

  async function handleDelete() {
    if (!doc) return;
    if (!confirm(`Delete "${friendlyUri(doc.file_uri)}"? This removes all indexed chunks.`)) return;
    setDeleting(true);
    try {
      await api.documents.delete(slug, id);
      router.push(`/collections/${slug}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Delete failed");
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
      })
      .catch(e => setError(e instanceof Error ? e.message : "Failed to load"))
      .finally(() => setLoading(false));
  }, [slug, id]);

  const totalPages = Math.ceil(chunks.length / PAGE_SIZE);
  const visibleChunks = chunks.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div>
        <div className="flex items-center gap-1.5 text-sm text-zinc-400 mb-2">
          <Link href="/collections" className="hover:text-zinc-600">Collections</Link>
          <ChevronRight className="w-3.5 h-3.5" />
          <Link href={`/collections/${slug}`} className="hover:text-zinc-600">{slug}</Link>
          <ChevronRight className="w-3.5 h-3.5" />
          <span className="text-zinc-700 truncate max-w-xs">
            {doc ? friendlyUri(doc.file_uri) : id}
          </span>
        </div>
        <div className="flex items-start justify-between gap-4">
          <h1 className="text-xl font-semibold truncate">
            {doc ? friendlyUri(doc.file_uri) : "Chunks"}
          </h1>
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
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}

      {/* Document details */}
      {doc && (
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
            <p className="text-zinc-700">{new Date(doc.created_at).toLocaleString()}</p>
          </div>
          {doc.error && (
            <div className="col-span-2 sm:col-span-4">
              <p className="text-xs text-red-400">{doc.error}</p>
            </div>
          )}
        </div>
      )}

      {/* Chunks */}
      {loading ? (
        <p className="text-sm text-zinc-400">Loading…</p>
      ) : chunks.length === 0 ? (
        <p className="text-sm text-zinc-400">No chunks found for this document.</p>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-zinc-500">{chunks.length} chunks</p>
            {totalPages > 1 && (
              <div className="flex items-center gap-3 text-sm text-zinc-500">
                <span>
                  {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, chunks.length)} of {chunks.length}
                </span>
                <div className="flex gap-1">
                  <Button variant="outline" size="sm" disabled={page === 0} onClick={() => setPage(p => p - 1)}>
                    <ChevronLeft className="w-4 h-4" />
                  </Button>
                  <Button variant="outline" size="sm" disabled={page >= totalPages - 1} onClick={() => setPage(p => p + 1)}>
                    <ChevronRight className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-3">
            {visibleChunks.map(ch => (
              <div key={ch.id} className="rounded-lg border bg-white p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-zinc-500">Chunk #{ch.chunk_index}</span>
                  <span className="font-mono text-xs text-zinc-300" title={ch.chunk_hash}>
                    {ch.chunk_hash.slice(0, 16)}…
                  </span>
                </div>
                {ch.chunk_header && (
                  <>
                    <div className="flex items-center gap-1.5 text-xs text-indigo-500 font-medium">
                      <Hash className="w-3 h-3 shrink-0" />
                      <span>{ch.chunk_header}</span>
                    </div>
                    <div className="border-t border-zinc-100" />
                  </>
                )}
                <p className="text-sm text-zinc-700 whitespace-pre-wrap break-words leading-relaxed">
                  {ch.chunk_text ?? <span className="italic text-zinc-300">no text stored</span>}
                </p>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between text-sm text-zinc-500">
              <span>
                {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, chunks.length)} of {chunks.length}
              </span>
              <div className="flex gap-1">
                <Button variant="outline" size="sm" disabled={page === 0} onClick={() => setPage(p => p - 1)}>
                  <ChevronLeft className="w-4 h-4" />
                </Button>
                <Button variant="outline" size="sm" disabled={page >= totalPages - 1} onClick={() => setPage(p => p + 1)}>
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
