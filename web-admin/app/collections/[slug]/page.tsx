"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { api, Collection, Document } from "@/lib/api";
import { DocumentsTable, PAGE_SIZE } from "./_components/documents-table";

export default function CollectionPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();

  const [collection, setCollection] = useState<Collection | null>(null);
  const [docs, setDocs] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [deleting, setDeleting] = useState(false);

  const loadDocs = useCallback(async (p = page) => {
    try {
      const data = await api.documents.list(slug, PAGE_SIZE, p * PAGE_SIZE);
      setDocs(data.documents ?? []);
      setTotal(data.total);
    } catch { /* non-fatal */ }
  }, [slug, page]);

  useEffect(() => {
    api.collections.get(slug).then(setCollection).catch(() => toast.error("Collection not found"));
    loadDocs(0);
  }, [slug]);

  useEffect(() => { loadDocs(page); }, [page]);

  const hasActive = docs.some(d => d.status === "ingesting");
  useEffect(() => {
    if (!hasActive) return;
    const id = setInterval(() => loadDocs(page), 3000);
    return () => clearInterval(id);
  }, [hasActive, page, loadDocs]);

  async function handleDelete() {
    if (!confirm(`Delete "${collection?.name}"? This removes all indexed data.`)) return;
    setDeleting(true);
    try {
      await api.collections.delete(slug);
      router.push("/collections");
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Delete failed");
      setDeleting(false);
    }
  }

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold">{collection?.name ?? slug}</h1>
          {collection && (
            <p className="text-sm text-zinc-500 mt-0.5">
              {collection.embed_model} · {collection.vector_dim}d
            </p>
          )}
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="text-red-500 hover:text-red-600 hover:bg-red-50"
          onClick={handleDelete}
          disabled={deleting}
        >
          <Trash2 className="w-4 h-4 mr-1.5" />
          {deleting ? "Deleting…" : "Delete collection"}
        </Button>
      </div>

      <DocumentsTable
        slug={slug}
        docs={docs}
        total={total}
        page={page}
        onPageChange={setPage}
        onReload={() => loadDocs(page)}
        onIngest={() => router.push(`/collections/${slug}/ingest`)}
      />
    </div>
  );
}
