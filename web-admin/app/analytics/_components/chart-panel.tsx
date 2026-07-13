"use client";

import { ReactNode } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

interface ChartPanelProps {
  title: string;
  description?: string;
  loading: boolean;
  isEmpty: boolean;
  chart: ReactNode;
  table: ReactNode;
}

// Every chart's card: title/description, a chart/table toggle (the
// accessibility twin every chart needs per the dataviz skill), and an empty
// state for a fresh install with no telemetry data yet. While a new date
// range loads, the previous render stays visible at reduced opacity rather
// than swapping to a skeleton, so the layout never jumps.
export function ChartPanel({ title, description, loading, isEmpty, chart, table }: ChartPanelProps) {
  if (isEmpty) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          {description && <CardDescription>{description}</CardDescription>}
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-48 text-sm text-muted-foreground">
            No data in this range yet.
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Tabs defaultValue="chart">
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-4">
            <div>
              <CardTitle>{title}</CardTitle>
              {description && <CardDescription>{description}</CardDescription>}
            </div>
            <TabsList>
              <TabsTrigger value="chart">Chart</TabsTrigger>
              <TabsTrigger value="table">Table</TabsTrigger>
            </TabsList>
          </div>
        </CardHeader>
        <CardContent className={cn("transition-opacity", loading && "opacity-50")}>
          <TabsContent value="chart">{chart}</TabsContent>
          <TabsContent value="table">{table}</TabsContent>
        </CardContent>
      </Card>
    </Tabs>
  );
}
