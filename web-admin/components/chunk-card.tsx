import { Hash } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

interface ChunkCardProps {
  chunkIndex: number;
  chunkText?: string | null;
  chunkHeader?: string | null;
  extraJSON?: string | null;
  metadata?: Record<string, unknown> | null;
  /** Source document name — shown in search/RAG context */
  source?: string;
  /** File URI shown below the text */
  fileUri?: string;
  /** Vector cosine similarity 0–100, shown in search/RAG context */
  similarity?: number;
  /** Raw BM25 score from the keyword/FTS pass (hybrid results only, unnormalized) */
  keywordScore?: number;
  /** Fused weighted-RRF score, normalized 0–100 (hybrid results only) */
  rrfScoreNormalized?: number;
  /** Raw fused weighted-RRF score, unclamped (hybrid results only) */
  rrfScore?: number;
  /** Truncated hash shown in the chunks detail view */
  chunkHash?: string;
}

export function ChunkCard({
  chunkIndex,
  chunkText,
  chunkHeader,
  extraJSON,
  metadata,
  source,
  fileUri,
  similarity,
  keywordScore,
  rrfScoreNormalized,
  rrfScore,
  chunkHash,
}: ChunkCardProps) {
  return (
    <Card>
      <CardContent className="space-y-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {source ? (
              <>
                <span className="text-sm font-medium text-foreground">{source}</span>
                <span className="text-xs text-muted-foreground">chunk {chunkIndex}</span>
              </>
            ) : (
              <span className="text-xs font-semibold text-muted-foreground">Chunk #{chunkIndex}</span>
            )}
          </div>
          <div className="text-right">
            {similarity != null ? (
              <>
                <span className="text-lg font-semibold text-status-success-text">{similarity.toFixed(1)}%</span>
                <p className="text-xs text-muted-foreground">vector similarity</p>
              </>
            ) : chunkHash ? (
              <span className="font-mono text-xs text-muted-foreground/60" title={chunkHash}>
                {chunkHash.slice(0, 16)}…
              </span>
            ) : null}
          </div>
        </div>

        {(keywordScore != null || rrfScoreNormalized != null) && (
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            {rrfScoreNormalized != null && (
              <span>
                RRF <span className="font-medium text-foreground">{rrfScoreNormalized.toFixed(1)}%</span>
                {rrfScore != null && (
                  <span className="text-muted-foreground/70"> ({rrfScore.toFixed(4)} raw)</span>
                )}
              </span>
            )}
            {keywordScore != null && (
              <span>
                BM25 <span className="font-medium text-foreground">{keywordScore.toFixed(2)}</span>
              </span>
            )}
          </div>
        )}

        {chunkHeader && (
          <>
            <div className="flex items-center gap-1.5 text-xs text-primary font-medium">
              <Hash className="w-3 h-3 shrink-0" />
              <span>{chunkHeader}</span>
            </div>
            <div className="border-t border-border" />
          </>
        )}

        <p className="text-sm text-foreground whitespace-pre-wrap break-words leading-relaxed">
          {chunkText ?? <span className="italic text-muted-foreground/60">no text stored</span>}
        </p>

        {fileUri && <p className="text-xs text-muted-foreground font-mono">{fileUri}</p>}

        {metadata && Object.keys(metadata).length > 0 && (
          <div className="flex flex-wrap gap-2">
            {Object.entries(metadata).map(([k, v]) => (
              <span key={k} className="inline-flex items-center gap-1 rounded-full border border-border bg-muted/40 px-2 py-0.5 text-xs text-muted-foreground">
                <span className="font-medium text-muted-foreground/70">{k}</span>
                {String(v)}
              </span>
            ))}
          </div>
        )}

        {extraJSON && (
          <pre className="text-xs font-mono bg-muted/40 border border-border rounded p-2 overflow-x-auto text-muted-foreground whitespace-pre-wrap break-all">
            {formatJSON(extraJSON)}
          </pre>
        )}
      </CardContent>
    </Card>
  );
}

function formatJSON(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}
