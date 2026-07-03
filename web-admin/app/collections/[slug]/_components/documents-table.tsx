"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Trash2, RefreshCw, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { TableCell, TableRow } from "@/components/ui/table";
import { DataTable } from "@/components/data-table";
import { Pagination } from "@/components/pagination";
import { api, Document } from "@/lib/api";
import { timeAgo } from "@/lib/utils";

const PAGE_SIZE = 20;

const statusColors: Record<string, string> = {
  complete:  "badge-success",
  ingesting: "badge-warning",
  failed:    "badge-error",
};

function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

interface DocumentsTableProps {
  slug: string;
  docs: Document[];
  total: number;
  page: number;
  onPageChange: (page: number) => void;
  onReload: () => void;
}

export { PAGE_SIZE };

export function DocumentsTable({ slug, docs, total, page, onPageChange, onReload }: DocumentsTableProps) {
  const router = useRouter();
  const [deletingDocId, setDeletingDocId] = useState<string | null>(null);
  const [refreshingDocId, setRefreshingDocId] = useState<string | null>(null);

  async function handleRefresh(doc: Document) {
    setRefreshingDocId(doc.id);
    try {
      await api.ingest.refresh(slug, { file_uri: doc.file_uri, mime_type: doc.mime_type });
      onReload();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Refresh failed");
    } finally {
      setRefreshingDocId(null);
    }
  }

  async function handleDelete(docId: string, label: string) {
    if (!confirm(`Delete "${label}"? This removes all indexed chunks for this document.`)) return;
    setDeletingDocId(docId);
    try {
      await api.documents.delete(slug, docId);
      onReload();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Delete failed");
    } finally {
      setDeletingDocId(null);
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-medium">Documents</h2>
          {total > 0 && <p className="text-xs text-zinc-400 mt-0.5">{total} total</p>}
        </div>
        <button
          onClick={onReload}
          className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700"
        >
          <RefreshCw className="w-3.5 h-3.5" /> Refresh
        </button>
      </div>

      <DataTable columns={[
        { label: "File" },
        { label: "Type" },
        { label: "Chunks" },
        { label: "Status" },
        { label: "Ingested" },
        { label: "", className: "w-20" },
      ]}>
        {docs.length === 0 ? (
          <TableRow>
            <TableCell colSpan={6} className="text-center text-zinc-400 py-10">
              No documents yet.
            </TableCell>
          </TableRow>
        ) : docs.map(d => (
          <TableRow
            key={d.id}
            className="cursor-pointer hover:bg-zinc-50"
            onClick={() => router.push(`/collections/${slug}/documents/${d.id}/chunks`)}
          >
            <TableCell className="font-mono text-xs text-zinc-600 max-w-xs truncate">
              {friendlyUri(d.file_uri)}
            </TableCell>
            <TableCell className="text-xs text-zinc-500">{d.mime_type}</TableCell>
            <TableCell className="text-xs text-zinc-500">{d.chunk_count}</TableCell>
            <TableCell>
              <div className="flex items-center gap-1.5">
                {d.status === "ingesting" && (
                  <Loader2 className="w-3 h-3 animate-spin text-amber-500 shrink-0" />
                )}
                <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[d.status] ?? ""}`}>
                  {d.status}
                </span>
              </div>
              {d.error && (
                <p className="text-xs text-red-400 mt-0.5 max-w-xs truncate" title={d.error}>{d.error}</p>
              )}
            </TableCell>
            <TableCell className="text-xs text-zinc-400" title={new Date(d.created_at).toLocaleString()}>
              {timeAgo(d.created_at)}
            </TableCell>
            <TableCell>
              <div className="flex items-center gap-2">
                {!d.file_uri.startsWith("upload://") && (
                  <button
                    onClick={e => { e.stopPropagation(); handleRefresh(d); }}
                    disabled={refreshingDocId === d.id || deletingDocId === d.id}
                    className="text-zinc-300 hover:text-blue-500 transition-colors disabled:opacity-40"
                    title="Re-ingest document"
                  >
                    <RefreshCw className={`w-4 h-4 ${refreshingDocId === d.id ? "animate-spin" : ""}`} />
                  </button>
                )}
                <button
                  onClick={e => { e.stopPropagation(); handleDelete(d.id, friendlyUri(d.file_uri)); }}
                  disabled={deletingDocId === d.id || refreshingDocId === d.id}
                  className="text-zinc-300 hover:text-red-500 transition-colors disabled:opacity-40"
                  title="Delete document"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </DataTable>

      <Pagination
        page={page}
        totalPages={totalPages}
        total={total}
        pageSize={PAGE_SIZE}
        onPageChange={onPageChange}
      />
    </div>
  );
}
