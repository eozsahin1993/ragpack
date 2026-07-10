"use client";

import { Trash2, RefreshCw, Loader2, Pencil } from "lucide-react";
import { TableCell, TableRow } from "@/components/ui/table";
import { DataTable } from "@/components/data-table";
import { Pagination } from "@/components/pagination";
import { Document } from "@/lib/api";
import { timeAgo, friendlyUri, friendlyMimeType } from "@/lib/utils";

const STATUS_COLORS: Record<string, string> = {
  complete:  "badge-success",
  ingesting: "badge-warning",
  failed:    "badge-error",
};

export function docLabel(doc: Document) {
  return doc.name ?? friendlyUri(doc.file_uri);
}

interface DocumentsTableProps {
  docs: Document[];
  onRowClick?: (doc: Document) => void;
  loading?: boolean;
  emptyMessage?: string;

  // optional columns
  showCollection?: boolean;
  getCollectionLabel?: (doc: Document) => string;
  showType?: boolean;
  showChunks?: boolean;
  dateField?: "created_at" | "updated_at";
  dateLabel?: string;

  // actions — each is only rendered when its handler is provided
  onEdit?: (doc: Document) => void;
  onRefresh?: (doc: Document) => void;
  onDelete?: (doc: Document) => void;
  editingId?: string | null;
  refreshingId?: string | null;
  deletingId?: string | null;

  // optional pagination — omit for compact views
  page?: number;
  totalPages?: number;
  total?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
}

export function DocumentsTable({
  docs,
  onRowClick,
  loading,
  emptyMessage = "No documents yet.",
  showCollection,
  getCollectionLabel,
  showType,
  showChunks,
  dateField = "created_at",
  dateLabel = dateField === "updated_at" ? "Updated" : "Ingested",
  onEdit,
  onRefresh,
  onDelete,
  editingId,
  refreshingId,
  deletingId,
  page,
  totalPages,
  total,
  pageSize,
  onPageChange,
}: DocumentsTableProps) {
  const showActions = !!(onEdit || onRefresh || onDelete);
  const busy = (id: string) => editingId === id || refreshingId === id || deletingId === id;

  const columns = [
    { label: "File" },
    ...(showCollection ? [{ label: "Collection" }] : []),
    ...(showType ? [{ label: "Type" }] : []),
    ...(showChunks ? [{ label: "Chunks" }] : []),
    { label: "Status" },
    { label: dateLabel },
    ...(showActions ? [{ label: "", className: "w-20" }] : []),
  ];

  const colSpan = columns.length;

  return (
    <div className="space-y-3">
      <DataTable columns={columns}>
        {loading ? (
          <TableRow>
            <TableCell colSpan={colSpan} className="text-center text-zinc-400 py-10">Loading…</TableCell>
          </TableRow>
        ) : docs.length === 0 ? (
          <TableRow>
            <TableCell colSpan={colSpan} className="text-center text-zinc-400 py-10">{emptyMessage}</TableCell>
          </TableRow>
        ) : docs.map(d => {
          return (
            <TableRow
              key={d.id}
              className={onRowClick ? "cursor-pointer hover:bg-zinc-50" : undefined}
              onClick={() => onRowClick?.(d)}
            >
              <TableCell className="max-w-xs">
                <p className="text-xs text-zinc-700 truncate">{docLabel(d)}</p>
                {d.name && <p className="text-[10px] text-zinc-400 truncate mt-0.5 font-mono">{friendlyUri(d.file_uri)}</p>}
              </TableCell>

              {showCollection && (
                <TableCell className="text-xs text-zinc-500">
                  {getCollectionLabel?.(d) ?? d.collection_id.slice(0, 8) + "…"}
                </TableCell>
              )}

              {showType && (
                <TableCell className="text-xs text-zinc-500">{friendlyMimeType(d.mime_type)}</TableCell>
              )}

              {showChunks && (
                <TableCell className="text-xs text-zinc-500">{d.chunk_count}</TableCell>
              )}

              <TableCell>
                <div className="flex items-center gap-1.5">
                  {d.status === "ingesting" && (
                    <Loader2 className="w-3 h-3 animate-spin text-amber-500 shrink-0" />
                  )}
                  <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${STATUS_COLORS[d.status] ?? ""}`}>
                    {d.status}
                  </span>
                </div>
                {d.error && (
                  <p className="text-xs text-red-400 mt-0.5 max-w-xs truncate" title={d.error}>{d.error}</p>
                )}
              </TableCell>

              <TableCell className="text-xs text-zinc-400" title={new Date(d[dateField]).toLocaleString()}>
                {timeAgo(d[dateField])}
              </TableCell>

              {showActions && (
                <TableCell>
                  <div className="flex items-center gap-2">
                    {onEdit && (
                      <button
                        onClick={e => { e.stopPropagation(); onEdit(d); }}
                        disabled={busy(d.id)}
                        className="text-zinc-300 hover:text-primary transition-colors disabled:opacity-40"
                        title="Edit document"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                    )}
                    {onRefresh && !d.file_uri.startsWith("upload://") && (
                      <button
                        onClick={e => { e.stopPropagation(); onRefresh(d); }}
                        disabled={busy(d.id)}
                        className="text-zinc-300 hover:text-primary transition-colors disabled:opacity-40"
                        title="Re-ingest document"
                      >
                        <RefreshCw className={`w-4 h-4 ${refreshingId === d.id ? "animate-spin" : ""}`} />
                      </button>
                    )}
                    {onDelete && (
                      <button
                        onClick={e => { e.stopPropagation(); onDelete(d); }}
                        disabled={busy(d.id)}
                        className="text-zinc-300 hover:text-red-500 transition-colors disabled:opacity-40"
                        title="Delete document"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </TableCell>
              )}
            </TableRow>
          );
        })}
      </DataTable>

      {onPageChange && totalPages !== undefined && total !== undefined && pageSize !== undefined && page !== undefined && (
        <Pagination
          page={page}
          totalPages={totalPages}
          total={total}
          pageSize={pageSize}
          onPageChange={onPageChange}
        />
      )}
    </div>
  );
}
