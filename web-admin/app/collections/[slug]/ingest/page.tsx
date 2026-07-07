"use client";

import { useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Upload, CheckCircle2, XCircle, Loader2, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";

type Mode = "uri" | "file";
type FileStatus = "pending" | "uploading" | "done" | "error";

interface FileItem {
  file: File;
  status: FileStatus;
  error?: string;
}

export default function IngestPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [mode, setMode] = useState<Mode>("uri");
  const [uri, setUri] = useState("");
  const [files, setFiles] = useState<FileItem[]>([]);
  const [extraJSON, setExtraJSON] = useState("");
  const [jsonError, setJsonError] = useState(false);
  const [ingesting, setIngesting] = useState(false);

  function handleExtraJSONChange(value: string) {
    setExtraJSON(value);
    if (value.trim() === "") {
      setJsonError(false);
    } else {
      try {
        JSON.parse(value);
        setJsonError(false);
      } catch {
        setJsonError(true);
      }
    }
  }

  function handleFilesSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const selected = Array.from(e.target.files ?? []);
    if (fileInputRef.current) fileInputRef.current.value = "";
    setFiles(selected.map(file => ({ file, status: "pending" })));
  }

  function removeFile(index: number) {
    setFiles(prev => prev.filter((_, i) => i !== index));
  }

  async function handleIngest() {
    if (jsonError) return;
    const metadata = extraJSON.trim() || undefined;
    setIngesting(true);

    try {
      if (mode === "uri") {
        await api.ingest.uri(slug, { file_uri: uri, mime_type: "", extra_json: metadata });
        router.push(`/collections/${slug}`);
      } else {
        for (let i = 0; i < files.length; i++) {
          setFiles(prev => prev.map((item, idx) => idx === i ? { ...item, status: "uploading" } : item));
          try {
            await api.ingest.upload(slug, files[i].file, metadata);
            setFiles(prev => prev.map((item, idx) => idx === i ? { ...item, status: "done" } : item));
          } catch (err) {
            const msg = err instanceof Error ? err.message : "Upload failed";
            setFiles(prev => prev.map((item, idx) => idx === i ? { ...item, status: "error", error: msg } : item));
          }
        }
        router.push(`/collections/${slug}`);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ingest failed");
    } finally {
      setIngesting(false);
    }
  }

  const canSubmit =
    !jsonError &&
    !ingesting &&
    (mode === "uri" ? uri.trim() !== "" : files.length > 0);

  return (
    <div className="space-y-6">
      <div>
        <button
          onClick={() => router.push(`/collections/${slug}`)}
          className="flex items-center gap-1.5 text-sm text-zinc-400 hover:text-zinc-700 mb-4"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>
        <h1 className="text-xl font-semibold">Ingest documents</h1>
        <p className="text-sm text-zinc-500 mt-0.5">
          Add documents to <span className="font-medium text-zinc-700">{slug}</span>
        </p>
      </div>

      <div className="rounded-lg border border-border bg-card shadow-sm p-6 space-y-5">
        {/* Mode toggle */}
        <div className="flex rounded-md border border-border overflow-hidden w-fit">
          <button
            onClick={() => setMode("uri")}
            className={`px-4 py-1.5 text-sm font-medium transition-colors ${
              mode === "uri"
                ? "bg-primary text-primary-foreground"
                : "text-zinc-500 hover:text-zinc-700 hover:bg-zinc-50"
            }`}
          >
            From URL
          </button>
          <button
            onClick={() => setMode("file")}
            className={`px-4 py-1.5 text-sm font-medium transition-colors border-l border-border ${
              mode === "file"
                ? "bg-primary text-primary-foreground"
                : "text-zinc-500 hover:text-zinc-700 hover:bg-zinc-50"
            }`}
          >
            Upload file
          </button>
        </div>

        {/* Source input */}
        {mode === "uri" ? (
          <Input
            value={uri}
            onChange={e => setUri(e.target.value)}
            placeholder="https://… or s3://bucket/key"
          />
        ) : (
          <div className="space-y-2">
            <input
              ref={fileInputRef}
              type="file"
              multiple
              accept=".txt,.md,.markdown,.html,.htm,.pdf,.docx,.pptx,.xlsx,.csv,.json,.xml"
              className="hidden"
              onChange={handleFilesSelected}
            />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="w-full rounded-md border-2 border-dashed border-zinc-200 hover:border-zinc-300 py-6 flex flex-col items-center gap-2 text-zinc-400 hover:text-zinc-500 transition-colors"
            >
              <Upload className="w-5 h-5" />
              <span className="text-sm">Click to choose files</span>
            </button>
            {files.length > 0 && (
              <ul className="space-y-1.5">
                {files.map((item, i) => (
                  <li key={i} className="flex items-center gap-2 text-sm">
                    {item.status === "pending"   && <span className="w-4 h-4 rounded-full border-2 border-zinc-200 shrink-0" />}
                    {item.status === "uploading" && <Loader2 className="w-4 h-4 shrink-0 animate-spin text-zinc-400" />}
                    {item.status === "done"      && <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-500" />}
                    {item.status === "error"     && <XCircle className="w-4 h-4 shrink-0 text-red-400" />}
                    <span className={`flex-1 truncate ${item.status === "error" ? "text-red-500" : "text-zinc-600"}`}>
                      {item.file.name}
                    </span>
                    {item.error && <span className="text-xs text-red-400 shrink-0">{item.error}</span>}
                    {item.status === "pending" && (
                      <button onClick={() => removeFile(i)} className="text-zinc-300 hover:text-zinc-500 shrink-0">
                        <X className="w-3.5 h-3.5" />
                      </button>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        {/* Shared metadata */}
        <div>
          <label className="text-xs text-zinc-500 mb-1 block">Metadata <span className="text-zinc-400">(optional)</span></label>
          <textarea
            value={extraJSON}
            onChange={e => handleExtraJSONChange(e.target.value)}
            placeholder='{"author": "Alice", "department": "eng"}'
            rows={3}
            className={`w-full rounded-md border px-3 py-2 text-xs font-mono resize-none bg-white placeholder:text-zinc-400 focus:outline-none focus:ring-1 ${
              jsonError
                ? "border-red-300 focus:ring-red-300"
                : "border-zinc-200 focus:ring-zinc-300"
            }`}
          />
          {jsonError && <p className="text-xs text-red-500 mt-0.5">Must be valid JSON</p>}
        </div>

        {/* Submit */}
        <Button onClick={handleIngest} disabled={!canSubmit} className="w-full">
          {ingesting ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : null}
          {ingesting ? "Ingesting…" : "Ingest"}
        </Button>
      </div>
    </div>
  );
}
