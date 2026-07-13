"use client";

import { useMemo } from "react";
import { Line, LineChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent, type ChartConfig } from "@/components/ui/chart";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChartPanel } from "./chart-panel";
import { VolumePoint } from "@/lib/api";

const config = {
  ingestion: { label: "Ingestion", color: "var(--chart-1)" },
  query: { label: "Query", color: "var(--chart-2)" },
  rag: { label: "RAG", color: "var(--chart-3)" },
} satisfies ChartConfig;

interface VolumeChartProps {
  points: VolumePoint[];
  loading: boolean;
}

export function VolumeChart({ points, loading }: VolumeChartProps) {
  // Long format (one row per day+event_type) -> wide format (one row per
  // day, one column per series) for Recharts' multi-line shape.
  const rows = useMemo(() => {
    const byDay = new Map<string, { day: string; ingestion: number; query: number; rag: number }>();
    for (const p of points) {
      const row = byDay.get(p.day) ?? { day: p.day, ingestion: 0, query: 0, rag: 0 };
      row[p.event_type] = p.count;
      byDay.set(p.day, row);
    }
    return [...byDay.values()].sort((a, b) => a.day.localeCompare(b.day));
  }, [points]);

  return (
    <ChartPanel
      title="Usage volume"
      description="Ingestion and query/RAG activity per day"
      loading={loading}
      isEmpty={points.length === 0}
      chart={
        <ChartContainer config={config} className="aspect-auto h-64 w-full">
          <LineChart data={rows} margin={{ left: 8, right: 8 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="day" tickLine={false} axisLine={false} tickMargin={8} minTickGap={24} />
            <YAxis tickLine={false} axisLine={false} tickMargin={8} allowDecimals={false} width={32} />
            <ChartTooltip content={<ChartTooltipContent indicator="line" />} />
            <ChartLegend content={<ChartLegendContent />} />
            <Line dataKey="ingestion" stroke="var(--color-ingestion)" strokeWidth={2} dot={false} />
            <Line dataKey="query" stroke="var(--color-query)" strokeWidth={2} dot={false} />
            <Line dataKey="rag" stroke="var(--color-rag)" strokeWidth={2} dot={false} />
          </LineChart>
        </ChartContainer>
      }
      table={
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Day</TableHead>
              <TableHead className="text-right">Ingestion</TableHead>
              <TableHead className="text-right">Query</TableHead>
              <TableHead className="text-right">RAG</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map(r => (
              <TableRow key={r.day}>
                <TableCell>{r.day}</TableCell>
                <TableCell className="text-right tabular-nums">{r.ingestion.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{r.query.toLocaleString()}</TableCell>
                <TableCell className="text-right tabular-nums">{r.rag.toLocaleString()}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      }
    />
  );
}
