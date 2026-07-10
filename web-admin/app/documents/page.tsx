"use client";

import { useEffect, useCallback, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { api, Document, Collection, MetadataField } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { DocumentsTable } from "@/components/documents/documents-table";
import { DocumentEditDialog } from "@/components/documents/document-edit-dialog";

const PAGE_SIZE = 50;

export default function DocumentsPage() {
  const router = useRouter();
  const [docs, setDocs] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [collections, setCollections] = useState<Record<string, { slug: string; name: string }>>({});
  const [loading, setLoading] = useState(true);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [refreshingId, setRefreshingId] = useState<string | null>(null);
  const [editingDoc, setEditingDoc] = useState<Document | null>(null);
  const [editingFields, setEditingFields] = useState<MetadataField[]>([]);
  const metadataFieldsCache = useRef<Record<string, MetadataField[]>>({});

  const fetchDocs = useCallback(async (showLoading = false, currentPage = page) => {
    if (showLoading) setLoading(true);
    try {
      const [docsData, collectionsData] = await Promise.all([
        api.documents.all(PAGE_SIZE, currentPage * PAGE_SIZE),
        api.collections.list(),
      ]);
      setDocs(docsData.documents ?? []);
      setTotal(docsData.total ?? 0);
      const colMap: Record<string, { slug: string; name: string }> = {};
      for (const c of (collectionsData.collections ?? []) as Collection[]) {
        colMap[c.id] = { slug: c.slug, name: c.name };
      }
      setCollections(colMap);
    } catch (e: unknown) {
      if (showLoading) toast.error(e instanceof Error ? e.message : "Failed to load documents");
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [page]);

  function handlePageChange(newPage: number) {
    setPage(newPage);
    fetchDocs(true, newPage);
  }

  async function handleDelete(doc: Document) {
    const col = collections[doc.collection_id];
    if (!col) return;
    if (!confirm(`Delete "${doc.name ?? doc.file_uri}"? This removes all indexed chunks.`)) return;
    setDeletingId(doc.id);
    try {
      await api.documents.delete(col.slug, doc.id);
      setDocs(prev => prev.filter(d => d.id !== doc.id));
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to delete document");
    } finally {
      setDeletingId(null);
    }
  }

  async function handleRefresh(doc: Document) {
    const col = collections[doc.collection_id];
    if (!col) return;
    setRefreshingId(doc.id);
    try {
      await api.ingest.refresh(col.slug, { file_uri: doc.file_uri, mime_type: doc.mime_type });
      fetchDocs(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Refresh failed");
    } finally {
      setRefreshingId(null);
    }
  }

  async function handleEdit(doc: Document) {
    const col = collections[doc.collection_id];
    if (!col) return;
    let fields = metadataFieldsCache.current[col.slug];
    if (!fields) {
      try {
        fields = (await api.metadataFields.list(col.slug)).fields ?? [];
        metadataFieldsCache.current[col.slug] = fields;
      } catch {
        fields = [];
      }
    }
    setEditingFields(fields);
    setEditingDoc(doc);
  }

  useEffect(() => { fetchDocs(true, 0); }, [fetchDocs]); // eslint-disable-line react-hooks/exhaustive-deps

  const hasActive = docs.some(d => d.status === "ingesting");
  useEffect(() => {
    if (!hasActive) return;
    const id = setInterval(() => fetchDocs(false), 1000);
    return () => clearInterval(id);
  }, [hasActive, fetchDocs]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Documents"
        description={total > 0 ? `${total} total, most recently updated first` : "All ingested documents across collections"}
        action={
          <button onClick={() => fetchDocs(true, page)} className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700">
            <RefreshCw className="w-3.5 h-3.5" /> Refresh
          </button>
        }
      />

      <DocumentsTable
        docs={docs}
        onRowClick={doc => router.push(`/documents/${doc.id}`)}
        showCollection
        getCollectionLabel={doc => collections[doc.collection_id]?.name ?? doc.collection_id.slice(0, 8) + "…"}
        showType
        dateField="updated_at"
        loading={loading}
        deletingId={deletingId}
        refreshingId={refreshingId}
        editingId={editingDoc?.id ?? null}
        onDelete={handleDelete}
        onRefresh={handleRefresh}
        onEdit={handleEdit}
        page={page}
        totalPages={Math.ceil(total / PAGE_SIZE)}
        total={total}
        pageSize={PAGE_SIZE}
        onPageChange={handlePageChange}
      />

      {editingDoc && (
        <DocumentEditDialog
          slug={collections[editingDoc.collection_id]?.slug ?? ""}
          doc={editingDoc}
          metadataFields={editingFields}
          onClose={() => setEditingDoc(null)}
          onSaved={() => {
            setEditingDoc(null);
            fetchDocs(false);
          }}
        />
      )}
    </div>
  );
}
