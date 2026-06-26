import { Hash } from "lucide-react";

interface ChunkCardProps {
  chunkIndex: number;
  chunkText?: string | null;
  chunkHeader?: string | null;
  /** Source document name — shown in search/RAG context */
  source?: string;
  /** File URI shown below the text */
  fileUri?: string;
  /** Similarity score 0–100, shown in search/RAG context */
  similarity?: number;
  /** Truncated hash shown in the chunks detail view */
  chunkHash?: string;
}

export function ChunkCard({
  chunkIndex,
  chunkText,
  chunkHeader,
  source,
  fileUri,
  similarity,
  chunkHash,
}: ChunkCardProps) {
  return (
    <div className="rounded-lg border bg-white p-5 space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {source ? (
            <>
              <span className="text-sm font-medium">{source}</span>
              <span className="text-xs text-zinc-400">chunk {chunkIndex}</span>
            </>
          ) : (
            <span className="text-xs font-semibold text-zinc-500">Chunk #{chunkIndex}</span>
          )}
        </div>
        <div className="text-right">
          {similarity != null ? (
            <>
              <span className="text-lg font-semibold text-emerald-600">{similarity.toFixed(1)}%</span>
              <p className="text-xs text-zinc-400">similarity</p>
            </>
          ) : chunkHash ? (
            <span className="font-mono text-xs text-zinc-300" title={chunkHash}>
              {chunkHash.slice(0, 16)}…
            </span>
          ) : null}
        </div>
      </div>

      {chunkHeader && (
        <>
          <div className="flex items-center gap-1.5 text-xs text-indigo-500 font-medium">
            <Hash className="w-3 h-3 shrink-0" />
            <span>{chunkHeader}</span>
          </div>
          <div className="border-t border-zinc-100" />
        </>
      )}

      <p className="text-sm text-zinc-700 whitespace-pre-wrap break-words leading-relaxed">
        {chunkText ?? <span className="italic text-zinc-300">no text stored</span>}
      </p>

      {fileUri && <p className="text-xs text-zinc-400 font-mono">{fileUri}</p>}
    </div>
  );
}
