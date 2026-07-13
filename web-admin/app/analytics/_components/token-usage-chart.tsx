"use client";

import { type ChartConfig } from "@/components/ui/chart";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChartPanel } from "./chart-panel";
import { GroupedBarChart } from "@/components/charts/grouped-bar-chart";
import { CollectionTokens } from "@/lib/api";

// 4 series -- the CVD floor per the dataviz skill's series-count ladder, so
// this chart leans on the legend + gaps + table (never color alone) rather
// than dropping any series.
const config = {
  ingestion_embed_tokens: { label: "Ingestion embed", color: "var(--chart-1)" },
  query_embed_tokens: { label: "Query embed", color: "var(--chart-2)" },
  llm_input_tokens: { label: "LLM input", color: "var(--chart-3)" },
  llm_output_tokens: { label: "LLM output", color: "var(--chart-4)" },
} satisfies ChartConfig;

interface TokenUsageChartProps {
  collections: CollectionTokens[];
  loading: boolean;
}

export function TokenUsageChart({ collections, loading }: TokenUsageChartProps) {
  return (
    <ChartPanel
      title="Token usage by collection"
      description="Embedding and LLM tokens"
      loading={loading}
      isEmpty={collections.length === 0}
      chart={
        <GroupedBarChart
          data={collections}
          categoryKey="collection_slug"
          seriesKeys={["ingestion_embed_tokens", "query_embed_tokens", "llm_input_tokens", "llm_output_tokens"]}
          config={config}
        />
      }
      table={
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Collection</TableHead>
              <TableHead className="text-right">Ingestion embed</TableHead>
              <TableHead className="text-right">Query embed</TableHead>
              <TableHead className="text-right">LLM input</TableHead>
              <TableHead className="text-right">LLM output</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {collections.map(c => (
              <TableRow key={c.collection_slug}>
                <TableCell>{c.collection_slug}</TableCell>
                <TableCell className="text-right tabular-nums">{c.ingestion_embed_tokens.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{c.query_embed_tokens.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{c.llm_input_tokens.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{c.llm_output_tokens.toLocaleString()}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      }
    />
  );
}
