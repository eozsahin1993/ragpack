"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { PageHeader } from "@/components/page-header";
import { api, VolumePoint, CollectionCost, LatencyBucket, MimeSuccessRate, CollectionTokens } from "@/lib/api";
import { DaysSelect } from "./_components/days-select";
import { VolumeChart } from "./_components/volume-chart";
import { CostChart } from "./_components/cost-chart";
import { LatencyChart } from "./_components/latency-chart";
import { SuccessRateChart } from "./_components/success-rate-chart";
import { TokenUsageChart } from "./_components/token-usage-chart";

export default function AnalyticsPage() {
  const [days, setDays] = useState(30);
  const [loading, setLoading] = useState(true);
  const [volume, setVolume] = useState<VolumePoint[]>([]);
  const [cost, setCost] = useState<CollectionCost[]>([]);
  const [latency, setLatency] = useState<LatencyBucket[]>([]);
  const [successRate, setSuccessRate] = useState<MimeSuccessRate[]>([]);
  const [tokens, setTokens] = useState<CollectionTokens[]>([]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([
      api.analytics.volume(days),
      api.analytics.costByCollection(days),
      api.analytics.latency(days),
      api.analytics.ingestionSuccessRate(days),
      api.analytics.tokenUsage(days),
    ])
      .then(([v, c, l, s, t]) => {
        if (cancelled) return;
        setVolume(v.points ?? []);
        setCost(c.collections ?? []);
        setLatency(l.buckets ?? []);
        setSuccessRate(s.mime_types ?? []);
        setTokens(t.collections ?? []);
      })
      .catch((e: unknown) => {
        if (!cancelled) toast.error(e instanceof Error ? e.message : "Failed to load analytics");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [days]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Analytics"
        description="Usage, cost, and performance across your RagPack instance."
      />

      <div className="flex justify-start">
        <DaysSelect days={days} onChange={setDays} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <VolumeChart points={volume} loading={loading} />
        <CostChart collections={cost} loading={loading} />
        <LatencyChart buckets={latency} loading={loading} />
        <SuccessRateChart mimeTypes={successRate} loading={loading} />
        <TokenUsageChart collections={tokens} loading={loading} />
      </div>
    </div>
  );
}
