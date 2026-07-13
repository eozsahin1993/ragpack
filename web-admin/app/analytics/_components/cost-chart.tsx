"use client";

import { type ChartConfig } from "@/components/ui/chart";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChartPanel } from "./chart-panel";
import { GroupedBarChart } from "@/components/charts/grouped-bar-chart";
import { CollectionCost } from "@/lib/api";

const config = {
  ingestion_cost_usd: { label: "Ingestion", color: "var(--chart-1)" },
  llm_cost_usd: { label: "LLM", color: "var(--chart-2)" },
} satisfies ChartConfig;

interface CostChartProps {
  collections: CollectionCost[];
  loading: boolean;
}

const formatUSD = (v: number) => `$${v < 0.01 && v > 0 ? v.toFixed(4) : v.toFixed(2)}`;

export function CostChart({ collections, loading }: CostChartProps) {
  return (
    <ChartPanel
      title="Cost by collection"
      description="Embedding vs. LLM spend"
      loading={loading}
      isEmpty={collections.length === 0}
      chart={
        <GroupedBarChart
          data={collections}
          categoryKey="collection_slug"
          seriesKeys={["ingestion_cost_usd", "llm_cost_usd"]}
          config={config}
          valueFormatter={formatUSD}
          stacked
        />
      }
      table={
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Collection</TableHead>
              <TableHead className="text-right">Ingestion</TableHead>
              <TableHead className="text-right">LLM</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {collections.map(c => (
              <TableRow key={c.collection_slug}>
                <TableCell>{c.collection_slug}</TableCell>
                <TableCell className="text-right tabular-nums">{formatUSD(c.ingestion_cost_usd)}</TableCell>
                <TableCell className="text-right tabular-nums">{formatUSD(c.llm_cost_usd)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      }
    />
  );
}
