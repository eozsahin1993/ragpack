"use client";

import { useRef, useState } from "react";
import { Upload, CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";

type UploadStatus = "pending" | "uploading" | "done" | "error";

interface UploadItem {
  name: string;
  status: UploadStatus;
  error?: string;
}

interface FileUploadProps {
  slug: string;
  onComplete: () => void;
}

export function FileUpload({ slug, onComplete }: FileUploadProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [queue, setQueue] = useState<UploadItem[]>([]);

  async function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    if (files.length === 0) return;
    if (fileInputRef.current) fileInputRef.current.value = "";

    setQueue(files.map(f => ({ name: f.name, status: "pending" })));

    for (let i = 0; i < files.length; i++) {
      setQueue(q => q.map((item, idx) => idx === i ? { ...item, status: "uploading" } : item));
      try {
        await api.ingest.upload(slug, files[i]);
        setQueue(q => q.map((item, idx) => idx === i ? { ...item, status: "done" } : item));
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Upload failed";
        setQueue(q => q.map((item, idx) => idx === i ? { ...item, status: "error", error: msg } : item));
      }
    }

    onComplete();
  }

  const uploading = queue.some(f => f.status === "uploading");

  return (
    <div className="space-y-3">
      <input
        ref={fileInputRef}
        type="file"
        multiple
        accept=".txt,.md,.markdown,.html,.htm,.pdf,.docx,.pptx,.xlsx,.csv,.json,.xml"
        className="hidden"
        onChange={handleChange}
      />
      <Button
        type="button"
        disabled={uploading}
        onClick={() => fileInputRef.current?.click()}
        className="gap-2 w-full"
      >
        <Upload className="w-4 h-4" />
        Choose files to upload
      </Button>

      {queue.length > 0 && (
        <ul className="space-y-1.5">
          {queue.map((item, i) => (
            <li key={i} className="flex items-center gap-2 text-sm">
              {item.status === "pending"   && <span className="w-4 h-4 rounded-full border-2 border-zinc-200 shrink-0" />}
              {item.status === "uploading" && <Loader2 className="w-4 h-4 shrink-0 animate-spin text-zinc-400" />}
              {item.status === "done"      && <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-500" />}
              {item.status === "error"     && <XCircle className="w-4 h-4 shrink-0 text-red-400" />}
              <span className={`truncate ${item.status === "error" ? "text-red-500" : "text-zinc-600"}`}>
                {item.name}
              </span>
              {item.error && (
                <span className="text-xs text-red-400 shrink-0">{item.error}</span>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
