"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { api, Collection, Document } from "@/lib/api";
import { FileUpload } from "./_components/file-upload";
import { UriIngest } from "./_components/uri-ingest";
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

      <div className="rounded-lg border border-border bg-card shadow-sm p-6 space-y-3">
        <h2 className="text-sm font-medium">Ingest document</h2>
        <UriIngest slug={slug} onComplete={() => loadDocs(page)} />
        <div className="flex items-center gap-3">
          <div className="flex-1 border-t border-border" />
          <span className="text-xs text-muted-foreground">or upload files</span>
          <div className="flex-1 border-t border-border" />
        </div>
        <FileUpload slug={slug} onComplete={() => loadDocs(page)} />
      </div>

      <DocumentsTable
        slug={slug}
        docs={docs}
        total={total}
        page={page}
        onPageChange={setPage}
        onReload={() => loadDocs(page)}
      />
    </div>
  );
}
