"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { api, Chunk, Document } from "@/lib/api"; // Document used by useState type inference
import { Pagination } from "@/components/pagination";
import { ChunkCard } from "@/components/chunk-card";
import { DocumentDetails } from "./_components/document-details";

const PAGE_SIZE = 20;


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

      {error && <p className="text-sm text-red-500">{error}</p>}

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
