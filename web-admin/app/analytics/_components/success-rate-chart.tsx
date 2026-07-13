"use client";

import { type ChartConfig } from "@/components/ui/chart";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChartPanel } from "./chart-panel";
import { GroupedBarChart } from "@/components/charts/grouped-bar-chart";
import { MimeSuccessRate } from "@/lib/api";

const config = {
  success_rate: { label: "Success rate", color: "var(--chart-1)" },
} satisfies ChartConfig;

interface SuccessRateChartProps {
  mimeTypes: MimeSuccessRate[];
  loading: boolean;
}

const formatPct = (v: number) => `${(v * 100).toFixed(1)}%`;

export function SuccessRateChart({ mimeTypes, loading }: SuccessRateChartProps) {
  return (
    <ChartPanel
      title="Ingestion success rate"
      description="By file type"
      loading={loading}
      isEmpty={mimeTypes.length === 0}
      chart={
        <GroupedBarChart
          data={mimeTypes}
          categoryKey="mime_type"
          seriesKeys={["success_rate"]}
          config={config}
          valueFormatter={formatPct}
          valueDomain={[0, 1]}
        />
      }
      table={
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Mime type</TableHead>
              <TableHead className="text-right">Total</TableHead>
              <TableHead className="text-right">Success rate</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mimeTypes.map(m => (
              <TableRow key={m.mime_type}>
                <TableCell>{m.mime_type}</TableCell>
                <TableCell className="text-right tabular-nums">{m.total_count.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{formatPct(m.success_rate)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      }
    />
  );
}
