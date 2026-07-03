"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Database, FileText, BriefcaseBusiness } from "lucide-react";
import { timeAgo } from "@/lib/utils";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { TableCell, TableRow } from "@/components/ui/table";
import { DataTable } from "@/components/data-table";
import { StatCard } from "@/components/dashboard/stat-card";
import { HealthCard, HealthStatus } from "@/components/dashboard/health-card";
import { CollectionCard } from "@/components/dashboard/collection-card";
import { api, Collection, Job, HealthInfo } from "@/lib/api";

const jobStatusColors: Record<string, string> = {
  complete:   "badge-success",
  processing: "badge-warning",
  pending:    "badge-warning",
  failed:     "badge-error",
};

function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

interface CollectionWithDocs extends Collection {
  docCount: number | null;
}

function JobSection({ title, jobs, onViewAll }: { title: string; jobs: Job[]; onViewAll: () => void }) {
  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{title}</h2>
        <button className="text-xs text-muted-foreground hover:text-foreground" onClick={onViewAll}>
          View all →
        </button>
      </div>
      <DataTable columns={[{ label: "File" }, { label: "Status" }, { label: "Updated" }]}>
        {jobs.map(j => (
          <TableRow key={j.id}>
            <TableCell className="font-mono text-xs text-muted-foreground max-w-xs truncate">
              {friendlyUri(j.file_uri)}
            </TableCell>
            <TableCell>
              <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${jobStatusColors[j.status] ?? ""}`}>
                {j.status}
              </span>
              {j.error && (
                <p className="text-xs text-red-400 mt-0.5 max-w-xs truncate" title={j.error}>
                  {j.error}
                </p>
              )}
            </TableCell>
            <TableCell className="text-xs text-muted-foreground" title={new Date(j.updated_at).toLocaleString()}>
              {timeAgo(j.updated_at)}
            </TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}

export default function DashboardPage() {
  const router = useRouter();

  const [collections, setCollections] = useState<CollectionWithDocs[]>([]);
  const [recentJobs, setRecentJobs] = useState<Job[]>([]);
  const [recentFailedJobs, setRecentFailedJobs] = useState<Job[]>([]);
  const [backendHealth, setBackendHealth] = useState<HealthInfo | null>(null);
  const [backendStatus, setBackendStatus] = useState<HealthStatus>("loading");
  const [embedderStatus, setEmbedderStatus] = useState<HealthStatus>("loading");
  const [embedderModel, setEmbedderModel] = useState<string | undefined>();
  const [llmStatus, setLlmStatus] = useState<HealthStatus>("loading");
  const [llmModel, setLlmModel] = useState<string | undefined>();
  const [totalDocs, setTotalDocs] = useState<number | null>(null);
  const [activeJobCount, setActiveJobCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      // Collections + jobs in parallel
      const [collectionsRes, jobsRes] = await Promise.allSettled([
        api.collections.list(),
        api.jobs.all(),
      ]);

      const cols: CollectionWithDocs[] =
        collectionsRes.status === "fulfilled"
          ? (collectionsRes.value.collections ?? []).map(c => ({ ...c, docCount: null }))
          : [];

      // Doc counts for each collection (N+1 but unavoidable without a dedicated endpoint)
      const docCountResults = await Promise.allSettled(
        cols.map(c => api.documents.list(c.slug, 1, 0))
      );
      let total = 0;
      const colsWithDocs = cols.map((c, i) => {
        const count =
          docCountResults[i].status === "fulfilled"
            ? docCountResults[i].value.total
            : null;
        if (count != null) total += count;
        return { ...c, docCount: count };
      });
      setCollections(colsWithDocs);
      setTotalDocs(total);

      if (jobsRes.status === "fulfilled") {
        const jobs = jobsRes.value.jobs ?? [];
        setRecentJobs(jobs.slice(0, 5));
        setRecentFailedJobs(jobs.filter(j => j.status === "failed").slice(0, 5));
        setActiveJobCount(jobs.filter(j => j.status === "processing" || j.status === "pending").length);
      }

      setLoading(false);

      // Health checks run in parallel, update independently
      api.health
        .get()
        .then(d => { setBackendHealth(d); setBackendStatus("ok"); })
        .catch(() => setBackendStatus("error"));

      api.embedders
        .list()
        .then(d => {
          setEmbedderModel(d.default || d.models?.[0]);
          setEmbedderStatus("ok");
        })
        .catch(() => setEmbedderStatus("error"));

      api.llms
        .list()
        .then(d => {
          setLlmModel(d.default || d.models?.[0]);
          setLlmStatus("ok");
        })
        .catch(() => setLlmStatus("error"));
    }

    load();
  }, []);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold">Overview</h1>
        <p className="text-sm text-muted-foreground mt-0.5">
          System health and activity at a glance
        </p>
      </div>

      {/* Health cards */}
      <div>
        <h2 className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-3">
          System health
        </h2>
        <div className="grid grid-cols-3 gap-4">
          <HealthCard
            label="Backend"
            status={backendStatus}
            model={backendHealth ? `v${backendHealth.version} · up ${backendHealth.uptime}` : undefined}
            icon="backend"
          />
          <HealthCard
            label="Embedding provider"
            status={embedderStatus}
            model={embedderModel}
            icon="embedder"
          />
          <HealthCard
            label="LLM provider"
            status={llmStatus}
            model={llmModel}
            icon="llm"
            detail="Not configured — set LLM_PROVIDER in .env.ragpack"
          />
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-3 gap-4">
        <StatCard
          label="Collections"
          value={collections.length}
          icon={Database}
          loading={loading}
        />
        <StatCard
          label="Documents"
          value={totalDocs ?? "—"}
          icon={FileText}
          loading={loading}
        />
        <StatCard
          label="Active jobs"
          value={activeJobCount}
          icon={BriefcaseBusiness}
          loading={loading}
          accent={activeJobCount > 0 ? "amber" : "default"}
        />
      </div>

      {/* Collections grid */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Collections
          </h2>
          <button
            className="text-xs text-muted-foreground hover:text-foreground"
            onClick={() => router.push("/collections")}
          >
            View all →
          </button>
        </div>
        {loading ? (
          <div className="grid grid-cols-2 gap-4">
            {[0, 1].map(i => (
              <Card key={i} className="animate-pulse">
                <CardHeader>
                  <div className="h-4 bg-zinc-100 rounded w-1/2" />
                  <div className="h-3 bg-zinc-100 rounded w-1/3 mt-1" />
                </CardHeader>
                <CardContent>
                  <div className="h-3 bg-zinc-100 rounded w-1/4" />
                </CardContent>
              </Card>
            ))}
          </div>
        ) : collections.length === 0 ? (
          <Card>
            <CardContent className="py-10 text-center text-sm text-muted-foreground">
              No collections yet.{" "}
              <button
                className="text-primary hover:underline"
                onClick={() => router.push("/collections")}
              >
                Create your first collection
              </button>
              .
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-2 gap-4">
            {collections.map(c => (
              <CollectionCard key={c.id} collection={c} docCount={c.docCount} />
            ))}
          </div>
        )}
      </div>

      {/* Recent jobs */}
      {recentJobs.length > 0 && (
        <JobSection title="Recent jobs" jobs={recentJobs} onViewAll={() => router.push("/jobs")} />
      )}

      {/* Recent failed jobs */}
      {recentFailedJobs.length > 0 && (
        <JobSection title="Recent failed jobs" jobs={recentFailedJobs} onViewAll={() => router.push("/jobs")} />
      )}
    </div>
  );
}
