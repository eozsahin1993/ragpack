"use client";

import { Loader2, Trash2 } from "lucide-react";
import { TableCell, TableRow } from "@/components/ui/table";
import { DataTable } from "@/components/data-table";
import { Pagination } from "@/components/pagination";
import { Job } from "@/lib/api";
import { timeAgo, friendlyUri, friendlyMimeType } from "@/lib/utils";

const STATUS_COLORS: Record<string, string> = {
  complete:   "badge-success",
  processing: "badge-warning",
  pending:    "badge-warning",
  queued:     "badge-warning",
  failed:     "badge-error",
};

interface JobsTableProps {
  jobs: Job[];
  loading?: boolean;
  emptyMessage?: string;
  // optional columns
  showCollection?: boolean;
  showType?: boolean;
  showDelete?: boolean;
  collectionNames?: Record<string, string>;
  deletingId?: string | null;
  onDelete?: (job: Job) => void;
  // optional pagination — omit for compact views
  page?: number;
  totalPages?: number;
  total?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
}

export function JobsTable({
  jobs,
  loading,
  emptyMessage = "No jobs yet.",
  showCollection,
  showType,
  showDelete,
  collectionNames = {},
  deletingId,
  onDelete,
  page,
  totalPages,
  total,
  pageSize,
  onPageChange,
}: JobsTableProps) {
  const columns = [
    { label: "File" },
    ...(showCollection ? [{ label: "Collection" }] : []),
    ...(showType       ? [{ label: "Type" }]       : []),
    { label: "Status" },
    { label: "Created" },
    ...(showDelete     ? [{ label: "", className: "w-10" }] : []),
  ];

  const colSpan = columns.length;

  return (
    <div className="space-y-3">
      <DataTable columns={columns}>
        {loading ? (
          <TableRow>
            <TableCell colSpan={colSpan} className="text-center text-zinc-400 py-10">Loading…</TableCell>
          </TableRow>
        ) : jobs.length === 0 ? (
          <TableRow>
            <TableCell colSpan={colSpan} className="text-center text-zinc-400 py-10">{emptyMessage}</TableCell>
          </TableRow>
        ) : jobs.map(j => (
          <TableRow key={j.id} className="group">
            <TableCell className="max-w-xs">
              <p className="text-xs text-zinc-700 truncate">{j.display_name ?? friendlyUri(j.file_uri)}</p>
              {j.display_name && (
                <p className="text-[10px] text-zinc-400 truncate mt-0.5 font-mono">{friendlyUri(j.file_uri)}</p>
              )}
            </TableCell>

            {showCollection && (
              <TableCell className="text-xs text-zinc-500">
                {collectionNames[j.collection_id] ?? j.collection_id.slice(0, 8) + "…"}
              </TableCell>
            )}

            {showType && (
              <TableCell className="text-xs text-zinc-500">
                {j.intent ? (
                  <span className="px-2 py-0.5 rounded-full border font-medium bg-accent text-primary border-primary/20">
                    {j.intent}
                  </span>
                ) : (
                  friendlyMimeType(j.mime_type)
                )}
              </TableCell>
            )}

            <TableCell>
              <div className="flex items-center gap-1.5">
                {(j.status === "pending" || j.status === "processing") && (
                  <Loader2 className="w-3 h-3 animate-spin text-amber-500 shrink-0" />
                )}
                <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${STATUS_COLORS[j.status] ?? STATUS_COLORS.queued}`}>
                  {j.status}
                </span>
              </div>
              {j.error && (
                <p className="text-xs text-red-400 mt-0.5 max-w-xs truncate" title={j.error}>{j.error}</p>
              )}
            </TableCell>

            <TableCell className="text-xs text-zinc-400" title={new Date(j.created_at).toLocaleString()}>
              {timeAgo(j.created_at)}
            </TableCell>

            {showDelete && (
              <TableCell>
                {(j.status === "complete" || j.status === "failed") && onDelete && (
                  <button
                    onClick={() => onDelete(j)}
                    disabled={deletingId === j.id}
                    className="opacity-0 group-hover:opacity-100 transition-opacity text-zinc-300 hover:text-red-500 disabled:opacity-40"
                    title="Delete job"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </TableCell>
            )}
          </TableRow>
        ))}
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
