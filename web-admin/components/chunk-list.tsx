import { ChunkCard } from "@/components/chunk-card";

interface Chunk {
  id?: string;
  chunkIndex: number;
  chunkText?: string | null;
  chunkHeader?: string | null;
  source?: string;
  fileUri?: string;
  similarity?: number;
  chunkHash?: string;
}

interface ChunkListProps {
  chunks: Chunk[];
  emptyMessage?: string;
}

export function ChunkList({ chunks, emptyMessage = "No chunks found." }: ChunkListProps) {
  if (chunks.length === 0) {
    return <p className="text-sm text-zinc-400">{emptyMessage}</p>;
  }

  return (
    <div className="space-y-3">
      <p className="text-sm text-zinc-500">{chunks.length} chunk{chunks.length !== 1 ? "s" : ""}</p>
      {chunks.map((c, i) => (
        <ChunkCard
          key={c.id ?? i}
          chunkIndex={c.chunkIndex}
          chunkText={c.chunkText}
          chunkHeader={c.chunkHeader}
          source={c.source}
          fileUri={c.fileUri}
          similarity={c.similarity}
          chunkHash={c.chunkHash}
        />
      ))}
    </div>
  );
}
