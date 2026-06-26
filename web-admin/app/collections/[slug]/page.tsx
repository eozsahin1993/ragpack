"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Upload, Trash2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { api, Collection, Document } from "@/lib/api";
import { Pagination } from "@/components/pagination";

const MIME_TYPES = [
  { label: "Plain text (.txt)", value: "text/plain" },
  { label: "Markdown (.md)", value: "text/markdown" },
  { label: "HTML", value: "text/html" },
  { label: "PDF", value: "application/pdf" },
];

const statusColors: Record<string, string> = {
  complete:  "bg-emerald-50 text-emerald-700 border-emerald-200",
  ingesting: "bg-amber-50 text-amber-700 border-amber-200",
  failed:    "bg-red-50 text-red-700 border-red-200",
};

const PAGE_SIZE = 20;

function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

function guessMimeType(uri: string): string {
  const path = uri.toLowerCase().split("?")[0];
  if (path.endsWith(".pdf"))                          return "application/pdf";
  if (path.endsWith(".md") || path.endsWith(".markdown")) return "text/markdown";
  if (path.endsWith(".html") || path.endsWith(".htm")) return "text/html";
  if (path.endsWith(".txt"))                          return "text/plain";
  if (uri.startsWith("http://") || uri.startsWith("https://")) return "text/html";
  return "text/plain";
}

export default function CollectionPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();

  const [collection, setCollection] = useState<Collection | null>(null);
  const [docs, setDocs] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [ingestForm, setIngestForm] = useState({ file_uri: "", mime_type: "text/plain" });
  const [ingesting, setIngesting] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [deletingDocId, setDeletingDocId] = useState<string | null>(null);
  const [refreshingDocId, setRefreshingDocId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const loadDocs = useCallback(async (p = page) => {
    try {
      const data = await api.documents.list(slug, PAGE_SIZE, p * PAGE_SIZE);
      setDocs(data.documents ?? []);
      setTotal(data.total);
    } catch { /* non-fatal */ }
  }, [slug, page]);

  useEffect(() => {
    api.collections.get(slug).then(setCollection).catch(() => setError("Collection not found"));
    loadDocs(0);
  }, [slug]);

  useEffect(() => { loadDocs(page); }, [page]);

  async function handleIngest(e: React.FormEvent) {
    e.preventDefault();
    setIngesting(true);
    setError("");
    try {
      await api.ingest.uri(slug, { file_uri: ingestForm.file_uri, mime_type: ingestForm.mime_type });
      setIngestForm(f => ({ ...f, file_uri: "" }));
      await loadDocs(page);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Ingest failed");
    } finally {
      setIngesting(false);
    }
  }

  async function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setError("");
    try {
      await api.ingest.upload(slug, file);
      await loadDocs(page);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Upload failed");
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  async function handleRefreshDocument(doc: Document) {
    setRefreshingDocId(doc.id);
    setError("");
    try {
      await api.ingest.refresh(slug, { file_uri: doc.file_uri, mime_type: doc.mime_type });
      await loadDocs(page);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Refresh failed");
    } finally {
      setRefreshingDocId(null);
    }
  }

  async function handleDeleteDocument(docId: string, label: string) {
    if (!confirm(`Delete "${label}"? This removes all indexed chunks for this document.`)) return;
    setDeletingDocId(docId);
    try {
      await api.documents.delete(slug, docId);
      await loadDocs(page);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setDeletingDocId(null);
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

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-8">
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
              onChange={e => {
                const uri = e.target.value;
                setIngestForm(f => ({ ...f, file_uri: uri, mime_type: uri ? guessMimeType(uri) : f.mime_type }));
              }}
              placeholder="https://… or s3://bucket/key (server-side paths only)"
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

        <div className="flex items-center gap-3">
          <div className="flex-1 border-t border-zinc-100" />
          <span className="text-xs text-zinc-400">or upload a file</span>
          <div className="flex-1 border-t border-zinc-100" />
        </div>

        <input
          ref={fileInputRef}
          type="file"
          accept=".txt,.md,.html,.pdf"
          className="hidden"
          onChange={handleFileUpload}
        />
        <Button
          type="button"
          variant="outline"
          disabled={uploading}
          onClick={() => fileInputRef.current?.click()}
          className="gap-2 w-full"
        >
          <Upload className="w-4 h-4" />
          {uploading ? "Uploading…" : "Upload file (.txt, .md, .html, .pdf)"}
        </Button>
      </div>

      {/* Documents */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-medium">Documents</h2>
            {total > 0 && (
              <p className="text-xs text-zinc-400 mt-0.5">{total} total</p>
            )}
          </div>
          <button
            onClick={() => loadDocs(page)}
            className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-700"
          >
            <RefreshCw className="w-3.5 h-3.5" /> Refresh
          </button>
        </div>

        <div className="rounded-lg border bg-white overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow className="bg-zinc-50">
                <TableHead>File</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Chunks</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Ingested</TableHead>
                <TableHead className="w-20" />
              </TableRow>
            </TableHeader>
            <TableBody>
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
                  <TableCell className="text-xs text-zinc-500">
                    {d.chunk_count}
                  </TableCell>
                  <TableCell>
                    <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[d.status] ?? ""}`}>
                      {d.status}
                    </span>
                    {d.error && (
                      <p className="text-xs text-red-400 mt-0.5 max-w-xs truncate" title={d.error}>{d.error}</p>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-zinc-400">
                    {new Date(d.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {!d.file_uri.startsWith("upload://") && (
                        <button
                          onClick={e => { e.stopPropagation(); handleRefreshDocument(d); }}
                          disabled={refreshingDocId === d.id || deletingDocId === d.id}
                          className="text-zinc-300 hover:text-blue-500 transition-colors disabled:opacity-40"
                          title="Re-ingest document"
                        >
                          <RefreshCw className={`w-4 h-4 ${refreshingDocId === d.id ? "animate-spin" : ""}`} />
                        </button>
                      )}
                      <button
                        onClick={e => { e.stopPropagation(); handleDeleteDocument(d.id, friendlyUri(d.file_uri)); }}
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
            </TableBody>
          </Table>
        </div>

        <Pagination
          page={page}
          totalPages={totalPages}
          total={total}
          pageSize={PAGE_SIZE}
          onPageChange={setPage}
        />
      </div>
    </div>
  );
}
