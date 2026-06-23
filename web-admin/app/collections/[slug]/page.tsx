"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ChevronRight, Upload, Trash2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api, Collection, Job } from "@/lib/api";

const MIME_TYPES = [
  { label: "Plain text (.txt)", value: "text/plain" },
  { label: "Markdown (.md)", value: "text/markdown" },
  { label: "HTML", value: "text/html" },
  { label: "PDF", value: "application/pdf" },
];

const statusColors: Record<string, string> = {
  complete: "bg-emerald-50 text-emerald-700 border-emerald-200",
  processing: "bg-amber-50 text-amber-700 border-amber-200",
  failed: "bg-red-50 text-red-700 border-red-200",
  queued: "bg-zinc-100 text-zinc-600 border-zinc-200",
};

export default function CollectionPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();

  const [collection, setCollection] = useState<Collection | null>(null);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [ingestForm, setIngestForm] = useState({ file_uri: "", mime_type: "text/plain" });
  const [ingesting, setIngesting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");

  const loadJobs = useCallback(async () => {
    try {
      const data = await api.jobs.byCollection(slug);
      setJobs(data.jobs ?? []);
    } catch { /* non-fatal */ }
  }, [slug]);

  useEffect(() => {
    api.collections.get(slug).then(setCollection).catch(() => setError("Collection not found"));
    loadJobs();
  }, [slug, loadJobs]);

  async function handleIngest(e: React.FormEvent) {
    e.preventDefault();
    setIngesting(true);
    setError("");
    try {
      await api.ingest.uri(slug, { file_uri: ingestForm.file_uri, mime_type: ingestForm.mime_type });
      setIngestForm(f => ({ ...f, file_uri: "" }));
      await loadJobs();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Ingest failed");
    } finally {
      setIngesting(false);
    }
  }

  async function handleDelete() {
    if (!confirm(`Delete "${collection?.name}"? This removes all indexed data.`)) return;
    setDeleting(true);
    try {
      await api.collections.delete(slug);
      router.push("/collections");
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Delete failed");
      setDeleting(false);
    }
  }

  return (
    <div className="space-y-8">
      {/* Breadcrumb + header */}
      <div>
        <div className="flex items-center gap-1.5 text-sm text-zinc-400 mb-2">
          <Link href="/collections" className="hover:text-zinc-600">Collections</Link>
          <ChevronRight className="w-3.5 h-3.5" />
          <span className="text-zinc-700">{collection?.name ?? slug}</span>
        </div>
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-xl font-semibold">{collection?.name ?? slug}</h1>
            {collection && (
              <p className="text-sm text-zinc-500 mt-0.5">
                {collection.embed_model} · {collection.vector_dim}d
              </p>
            )}
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="text-red-500 hover:text-red-600 hover:bg-red-50"
            onClick={handleDelete}
            disabled={deleting}
          >
            <Trash2 className="w-4 h-4 mr-1.5" />
            {deleting ? "Deleting…" : "Delete collection"}
          </Button>
        </div>
        {error && <p className="text-red-500 text-sm mt-2">{error}</p>}
      </div>

      {/* Ingest */}
      <div className="rounded-lg border bg-white p-6 space-y-4">
        <div className="flex items-center gap-2">
          <Upload className="w-4 h-4 text-zinc-500" />
          <h2 className="font-medium">Ingest document</h2>
        </div>
        <form onSubmit={handleIngest} className="flex gap-3 items-end">
          <div className="flex-1 space-y-1.5">
            <Label className="text-xs text-zinc-500">File URI</Label>
            <Input
              required
              value={ingestForm.file_uri}
              onChange={e => setIngestForm(f => ({ ...f, file_uri: e.target.value }))}
              placeholder="https://… or file:///path/to/file.txt"
            />
          </div>
          <div className="w-52 space-y-1.5">
            <Label className="text-xs text-zinc-500">MIME type</Label>
            <Select
              value={ingestForm.mime_type}
              onValueChange={v => setIngestForm(f => ({ ...f, mime_type: v ?? f.mime_type }))}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {MIME_TYPES.map(m => (
                  <SelectItem key={m.value} value={m.value}>{m.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button type="submit" disabled={ingesting}>
            {ingesting ? "Ingesting…" : "Ingest"}
          </Button>
        </form>
      </div>

      {/* Documents / Jobs */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="font-medium">Indexed documents</h2>
          <button onClick={loadJobs} className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700">
            <RefreshCw className="w-3.5 h-3.5" /> Refresh
          </button>
        </div>
        <div className="rounded-lg border bg-white overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow className="bg-zinc-50">
                <TableHead>File URI</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Ingested</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {jobs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-zinc-400 py-10">
                    No documents ingested yet.
                  </TableCell>
                </TableRow>
              ) : jobs.map(j => (
                <TableRow key={j.id}>
                  <TableCell className="font-mono text-xs text-zinc-600 max-w-xs truncate">{j.file_uri}</TableCell>
                  <TableCell className="text-xs text-zinc-500">{j.mime_type}</TableCell>
                  <TableCell>
                    <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[j.status] ?? statusColors.queued}`}>
                      {j.status}
                    </span>
                  </TableCell>
                  <TableCell className="text-xs text-zinc-400">
                    {new Date(j.created_at).toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
