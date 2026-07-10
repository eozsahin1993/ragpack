"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Trash2, Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { api, Collection, Document, MetadataField } from "@/lib/api";
import { DocumentsTable, docLabel } from "@/components/documents/documents-table";
import { DocumentEditDialog } from "@/components/documents/document-edit-dialog";
import { MetadataFieldsPanel } from "./_components/metadata-fields-panel";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 20;

const TABS = [
  { key: "documents", label: "Documents" },
  { key: "metadata", label: "Document properties" },
] as const;
type TabKey = (typeof TABS)[number]["key"];

export default function CollectionPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();

  const [collection, setCollection] = useState<Collection | null>(null);
  const [docs, setDocs] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [metadataFields, setMetadataFields] = useState<MetadataField[]>([]);
  const [deleting, setDeleting] = useState(false);
  const [editingDoc, setEditingDoc] = useState<Document | null>(null);
  const [refreshingId, setRefreshingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabKey>("documents");

  const loadDocs = useCallback(async (p = page) => {
    try {
      const data = await api.documents.list(slug, PAGE_SIZE, p * PAGE_SIZE);
      setDocs(data.documents ?? []);
      setTotal(data.total);
    } catch { /* non-fatal */ }
  }, [slug, page]);

  useEffect(() => {
    api.collections.get(slug).then(setCollection).catch(() => toast.error("Collection not found"));
    api.metadataFields.list(slug).then(d => setMetadataFields(d.fields ?? [])).catch(() => {});
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

  async function handleRefreshDoc(doc: Document) {
    setRefreshingId(doc.id);
    try {
      await api.ingest.refresh(slug, { file_uri: doc.file_uri, mime_type: doc.mime_type });
      loadDocs(page);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Refresh failed");
    } finally {
      setRefreshingId(null);
    }
  }

  async function handleDeleteDoc(doc: Document) {
    if (!confirm(`Delete "${docLabel(doc)}"? This removes all indexed chunks for this document.`)) return;
    setDeletingId(doc.id);
    try {
      await api.documents.delete(slug, doc.id);
      loadDocs(page);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Delete failed");
    } finally {
      setDeletingId(null);
    }
  }

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold">{collection?.name ?? slug}</h1>
          {collection && (
            <p className="text-sm text-muted-foreground mt-0.5">
              {collection.embed_model} · {collection.vector_dim}d
            </p>
          )}
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive hover:text-destructive hover:bg-destructive/10"
          onClick={handleDelete}
          disabled={deleting}
        >
          <Trash2 className="w-4 h-4 mr-1.5" />
          {deleting ? "Deleting…" : "Delete collection"}
        </Button>
      </div>

      <div className="flex gap-1 border-b border-border">
        {TABS.map(t => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={cn(
              "px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
              activeTab === t.key
                ? "border-foreground text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            {t.label}
          </button>
        ))}
      </div>

      {activeTab === "documents" && (
        <div className="space-y-3 mt-4">
          <div className="flex items-center justify-between">
            {total > 0 ? <p className="text-xs text-muted-foreground">{total} total</p> : <span />}
            <Button size="sm" onClick={() => router.push(`/collections/${slug}/ingest`)}>
              <Plus className="w-4 h-4 mr-1" />
              Ingest
            </Button>
          </div>

          <DocumentsTable
            docs={docs}
            onRowClick={doc => router.push(`/collections/${slug}/documents/${doc.id}/chunks`)}
            showType
            showChunks
            dateField="created_at"
            onEdit={setEditingDoc}
            onRefresh={handleRefreshDoc}
            onDelete={handleDeleteDoc}
            refreshingId={refreshingId}
            deletingId={deletingId}
            page={page}
            totalPages={Math.ceil(total / PAGE_SIZE)}
            total={total}
            pageSize={PAGE_SIZE}
            onPageChange={setPage}
          />

          <DocumentEditDialog
            slug={slug}
            doc={editingDoc}
            metadataFields={metadataFields}
            onClose={() => setEditingDoc(null)}
            onSaved={() => {
              setEditingDoc(null);
              loadDocs(page);
            }}
          />
        </div>
      )}

      {activeTab === "metadata" && (
        <div className="mt-4">
          <MetadataFieldsPanel slug={slug} />
        </div>
      )}
    </div>
  );
}
