"use client";

import { useEffect, useState } from "react";
import { RefreshCw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import {
  TableCell,
  TableRow,
} from "@/components/ui/table";
import { api, Job, Collection } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { DataTable } from "@/components/data-table";

const statusColors: Record<string, string> = {
  complete: "badge-success",
  processing: "badge-warning",
  failed: "badge-error",
  queued: "badge-warning",
  pending: "badge-warning",
};

function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.floor(hrs / 24);
  if (days < 7) return `${days}d ago`;
  return new Date(dateStr).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

export default function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [collectionNames, setCollectionNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    try {
      const [jobsData, collectionsData] = await Promise.all([
        api.jobs.all(),
        api.collections.list(),
      ]);
      setJobs(jobsData.jobs ?? []);
      const nameMap: Record<string, string> = {};
      for (const c of (collectionsData.collections ?? []) as Collection[]) {
        nameMap[c.id] = c.name;
      }
      setCollectionNames(nameMap);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to load jobs");
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(job: Job) {
    setDeletingId(job.id);
    try {
      await api.jobs.delete(job.id);
      setJobs(prev => prev.filter(j => j.id !== job.id));
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to delete job");
    } finally {
      setDeletingId(null);
    }
  }

  useEffect(() => { load(); }, []);

  const canDelete = (j: Job) => j.status === "complete" || j.status === "failed";

  return (
    <div className="space-y-6">
      <PageHeader
        title="Jobs"
        description="All ingest jobs across collections"
        action={<button onClick={load} className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700"><RefreshCw className="w-3.5 h-3.5" /> Refresh</button>}
      />

      <DataTable columns={[
        { label: "File" },
        { label: "Collection" },
        { label: "Type" },
        { label: "Status" },
        { label: "Created" },
        { label: "", className: "w-10" },
      ]}>
        {loading ? (
          <TableRow>
            <TableCell colSpan={6} className="text-center text-zinc-400 py-10">Loading…</TableCell>
          </TableRow>
        ) : jobs.length === 0 ? (
          <TableRow>
            <TableCell colSpan={6} className="text-center text-zinc-400 py-10">No jobs yet.</TableCell>
          </TableRow>
        ) : jobs.map(j => (
          <TableRow key={j.id} className="group">
            <TableCell className="font-mono text-xs text-zinc-600 max-w-xs truncate">
              {friendlyUri(j.file_uri)}
            </TableCell>
            <TableCell className="text-xs text-zinc-500">
              {collectionNames[j.collection_id] ?? j.collection_id.slice(0, 8) + "…"}
            </TableCell>
            <TableCell className="text-xs text-zinc-500">{j.intent ?? j.mime_type}</TableCell>
            <TableCell>
              <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[j.status] ?? statusColors.queued}`}>
                {j.status}
              </span>
              {j.error && (
                <p className="text-xs text-red-400 mt-0.5 max-w-xs truncate" title={j.error}>{j.error}</p>
              )}
            </TableCell>
            <TableCell className="text-xs text-zinc-400" title={new Date(j.created_at).toLocaleString()}>{timeAgo(j.created_at)}</TableCell>
            <TableCell>
              {canDelete(j) && (
                <button
                  onClick={() => handleDelete(j)}
                  disabled={deletingId === j.id}
                  className="opacity-0 group-hover:opacity-100 transition-opacity text-zinc-300 hover:text-red-500 disabled:opacity-40"
                  title="Delete job"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              )}
            </TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}
