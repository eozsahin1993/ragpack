"use client";

import { type ChartConfig } from "@/components/ui/chart";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChartPanel } from "./chart-panel";
import { GroupedBarChart } from "@/components/charts/grouped-bar-chart";
import { LatencyBucket } from "@/lib/api";

const config = {
  p50_ms: { label: "p50", color: "var(--chart-1)" },
  p95_ms: { label: "p95", color: "var(--chart-2)" },
} satisfies ChartConfig;

interface LatencyChartProps {
  buckets: LatencyBucket[];
  loading: boolean;
}

const formatMs = (v: number) => `${Math.round(v).toLocaleString()}ms`;

export function LatencyChart({ buckets, loading }: LatencyChartProps) {
  return (
    <ChartPanel
      title="Query / RAG latency"
      description="p50 and p95 total response time, completed calls only"
      loading={loading}
      isEmpty={buckets.length === 0}
      chart={
        <GroupedBarChart
          data={buckets}
          categoryKey="endpoint"
          seriesKeys={["p50_ms", "p95_ms"]}
          config={config}
          valueFormatter={formatMs}
          valueAxisWidth={56}
        />
      }
      table={
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Endpoint</TableHead>
              <TableHead className="text-right">Samples</TableHead>
              <TableHead className="text-right">p50</TableHead>
              <TableHead className="text-right">p95</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {buckets.map(b => (
              <TableRow key={b.endpoint}>
                <TableCell className="capitalize">{b.endpoint}</TableCell>
                <TableCell className="text-right tabular-nums">{b.sample_count.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{formatMs(b.p50_ms)}</TableCell>
                <TableCell className="text-right tabular-nums">{formatMs(b.p95_ms)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      }
    />
  );
}
