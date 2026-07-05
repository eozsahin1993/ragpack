"use client";

import { useEffect, useCallback, useState } from "react";
import { RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { api, Job, Collection } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { JobsTable } from "@/components/jobs-table";

const PAGE_SIZE = 50;

export default function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [collectionNames, setCollectionNames] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchJobs = useCallback(async (showLoading = false, currentPage = page) => {
    if (showLoading) setLoading(true);
    try {
      const [jobsData, collectionsData] = await Promise.all([
        api.jobs.all(PAGE_SIZE, currentPage * PAGE_SIZE),
        api.collections.list(),
      ]);
      setJobs(jobsData.jobs ?? []);
      setTotal(jobsData.total ?? 0);
      const nameMap: Record<string, string> = {};
      for (const c of (collectionsData.collections ?? []) as Collection[]) {
        nameMap[c.id] = c.name;
      }
      setCollectionNames(nameMap);
    } catch (e: unknown) {
      if (showLoading) toast.error(e instanceof Error ? e.message : "Failed to load jobs");
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [page]);

  function handlePageChange(newPage: number) {
    setPage(newPage);
    fetchJobs(true, newPage);
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

  useEffect(() => { fetchJobs(true, 0); }, [fetchJobs]); // eslint-disable-line react-hooks/exhaustive-deps

  const hasActive = jobs.some(j => j.status === "pending" || j.status === "processing");
  useEffect(() => {
    if (!hasActive) return;
    const id = setInterval(() => fetchJobs(false), 1000);
    return () => clearInterval(id);
  }, [hasActive, fetchJobs]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Jobs"
        description="All ingest jobs across collections"
        action={
          <button onClick={() => fetchJobs(true, page)} className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700">
            <RefreshCw className="w-3.5 h-3.5" /> Refresh
          </button>
        }
      />

      <JobsTable
        jobs={jobs}
        loading={loading}
        showCollection
        showType
        showDelete
        collectionNames={collectionNames}
        deletingId={deletingId}
        onDelete={handleDelete}
        page={page}
        totalPages={Math.ceil(total / PAGE_SIZE)}
        total={total}
        pageSize={PAGE_SIZE}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
