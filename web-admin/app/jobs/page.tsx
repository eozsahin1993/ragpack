"use client";

import { useEffect, useState } from "react";
import { RefreshCw } from "lucide-react";
import { toast } from "sonner";
import {
  TableCell,
  TableRow,
} from "@/components/ui/table";
import { api, Job } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { DataTable } from "@/components/data-table";

const statusColors: Record<string, string> = {
  complete: "badge-success",
  processing: "badge-warning",
  failed: "badge-error",
  queued: "badge-warning",
};

export default function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    try {
      const data = await api.jobs.all();
      setJobs(data.jobs ?? []);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to load jobs");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Jobs"
        description="All ingest jobs across collections"
        action={<button onClick={load} className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700"><RefreshCw className="w-3.5 h-3.5" /> Refresh</button>}
      />

      <DataTable columns={[
        { label: "File URI" },
        { label: "Collection" },
        { label: "Type" },
        { label: "Status" },
        { label: "Created" },
      ]}>
        {loading ? (
          <TableRow>
            <TableCell colSpan={5} className="text-center text-zinc-400 py-10">Loading…</TableCell>
          </TableRow>
        ) : jobs.length === 0 ? (
          <TableRow>
            <TableCell colSpan={5} className="text-center text-zinc-400 py-10">No jobs yet.</TableCell>
          </TableRow>
        ) : jobs.map(j => (
          <TableRow key={j.id}>
            <TableCell className="font-mono text-xs text-zinc-600 max-w-xs truncate">{j.file_uri}</TableCell>
            <TableCell className="text-xs text-zinc-500">{j.collection_id}</TableCell>
            <TableCell className="text-xs text-zinc-500">{j.mime_type}</TableCell>
            <TableCell>
              <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[j.status] ?? statusColors.queued}`}>
                {j.status}
              </span>
            </TableCell>
            <TableCell className="text-xs text-zinc-400">{new Date(j.created_at).toLocaleString()}</TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}
