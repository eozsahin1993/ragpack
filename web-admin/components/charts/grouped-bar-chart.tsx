"use client";

import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent, type ChartConfig } from "@/components/ui/chart";

// Shared shape for "tell distinct series apart" across a few categories
// (categorical color, legend always present — see dataviz skill's
// choosing-a-form.md) — used by latency (2 series) and token usage (4
// series), which differ only in field names, series count, and formatting.
interface GroupedBarChartProps<T> {
  data: T[];
  categoryKey: keyof T & string;
  seriesKeys: (keyof T & string)[];
  config: ChartConfig;
  valueFormatter?: (v: number) => string;
  valueAxisWidth?: number;
  valueDomain?: [number, number];
  stacked?: boolean;
}

export function GroupedBarChart<T>({
  data,
  categoryKey,
  seriesKeys,
  config,
  valueFormatter,
  valueAxisWidth = 48,
  valueDomain,
  stacked = false,
}: GroupedBarChartProps<T>) {
  return (
    <ChartContainer config={config} className="aspect-auto h-64 w-full">
      <BarChart data={data} margin={{ left: 8, right: 8 }}>
        <CartesianGrid vertical={false} />
        <XAxis dataKey={categoryKey as string} tickLine={false} axisLine={false} tickMargin={8} />
        <YAxis domain={valueDomain} tickLine={false} axisLine={false} tickMargin={8} width={valueAxisWidth} tickFormatter={valueFormatter} />
        <ChartTooltip content={<ChartTooltipContent indicator="line" formatter={valueFormatter ? (v) => valueFormatter(Number(v)) : undefined} />} />
        {seriesKeys.length > 1 && <ChartLegend content={<ChartLegendContent />} />}
        {seriesKeys.map(key => (
          <Bar
            key={key as string}
            dataKey={key as string}
            fill={`var(--color-${key})`}
            radius={4}
            maxBarSize={20}
            stackId={stacked ? "stack" : undefined}
          />
        ))}
      </BarChart>
    </ChartContainer>
  );
}
