import { Document } from "@/lib/api";

const statusColors: Record<string, string> = {
  complete:  "bg-emerald-50 text-emerald-700 border-emerald-200",
  ingesting: "bg-amber-50 text-amber-700 border-amber-200",
  failed:    "bg-red-50 text-red-700 border-red-200",
};

interface DocumentDetailsProps {
  doc: Document;
}

export function DocumentDetails({ doc }: DocumentDetailsProps) {
  return (
    <div className="rounded-lg border bg-white p-4 grid grid-cols-2 gap-x-8 gap-y-2 text-sm sm:grid-cols-4">
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Type</p>
        <p className="text-zinc-700 font-mono text-xs">{doc.mime_type}</p>
      </div>
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Status</p>
        <span className={`text-xs px-2 py-0.5 rounded-full border font-medium ${statusColors[doc.status] ?? ""}`}>
          {doc.status}
        </span>
      </div>
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Chunks</p>
        <p className="text-zinc-700">{doc.chunk_count}</p>
      </div>
      <div>
        <p className="text-xs text-zinc-400 mb-0.5">Ingested</p>
        <p className="text-zinc-700">{new Date(doc.created_at).toLocaleString()}</p>
      </div>
      {doc.error && (
        <div className="col-span-2 sm:col-span-4">
          <p className="text-xs text-red-400">{doc.error}</p>
        </div>
      )}
    </div>
  );
}
