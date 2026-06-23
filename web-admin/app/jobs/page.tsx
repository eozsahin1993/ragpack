"use client";

import { useEffect, useState } from "react";
import { RefreshCw } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api, Job } from "@/lib/api";

const statusColors: Record<string, string> = {
  complete: "bg-emerald-50 text-emerald-700 border-emerald-200",
  processing: "bg-amber-50 text-amber-700 border-amber-200",
  failed: "bg-red-50 text-red-700 border-red-200",
  queued: "bg-zinc-100 text-zinc-600 border-zinc-200",
};

export default function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    try {
      const data = await api.jobs.all();
      setJobs(data.jobs ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Jobs</h1>
          <p className="text-sm text-zinc-500 mt-0.5">All ingest jobs across collections</p>
        </div>
        <button onClick={load} className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700">
          <RefreshCw className="w-3.5 h-3.5" /> Refresh
        </button>
      </div>

      {error && <p className="text-red-500 text-sm">{error}</p>}

      <div className="rounded-lg border bg-white overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="bg-zinc-50">
              <TableHead>File URI</TableHead>
              <TableHead>Collection</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
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
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
