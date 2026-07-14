"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { api, Collection } from "@/lib/api";
import { describeInterval } from "@/lib/utils";

// Mirrors backend/pkg/config/defaults.go's DefaultMinCollectionRefreshSeconds.
// The operator can override the real minimum via MIN_COLLECTION_REFRESH,
// which the backend enforces regardless of what's shown here — a save
// that's too low surfaces the backend's own error message, which is always
// the authoritative value. Hours-only input: below-an-hour granularity has
// no valid range left once the minimum itself is an hour, and days are just
// a multiple of hours, so a unit picker would be complexity with no payoff.
const MIN_INTERVAL_HOURS = 1;
const DEFAULT_INTERVAL_HOURS = 24;

interface Props {
  collection: Collection;
  onUpdated: (collection: Collection) => void;
}

// Status (enabled, interval, last checked) is shown in the page header —
// always visible regardless of tab. This panel is edit-only.
export function RefreshSettingsPanel({ collection, onUpdated }: Props) {
  const savedHours = collection.refresh_interval_seconds
    ? Math.round(collection.refresh_interval_seconds / 3600)
    : DEFAULT_INTERVAL_HOURS;

  const [enabled, setEnabled] = useState(collection.refresh_enabled ?? false);
  const [hours, setHours] = useState(String(savedHours));
  const [saving, setSaving] = useState(false);

  const dirty = enabled !== (collection.refresh_enabled ?? false) || (enabled && Number(hours) !== savedHours);

  async function handleSave() {
    const h = Number(hours);
    if (enabled && (!Number.isFinite(h) || h < MIN_INTERVAL_HOURS)) {
      toast.error(`Interval must be at least ${MIN_INTERVAL_HOURS} hour`);
      return;
    }
    setSaving(true);
    try {
      const updated = await api.collections.update(collection.id, {
        refresh_enabled: enabled,
        ...(enabled ? { refresh_interval_seconds: h * 3600 } : {}),
      });
      onUpdated(updated);
      toast.success(enabled ? "Auto-refresh enabled" : "Auto-refresh disabled");
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to update auto-refresh settings");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Auto-refresh</CardTitle>
        <CardDescription>
          Periodically re-check this collection&apos;s documents at their source and reingest anything that changed.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="flex items-center justify-between gap-4">
          <div className="space-y-0.5">
            <Label htmlFor="auto-refresh-toggle">Enable auto-refresh</Label>
            <p className="text-xs text-muted-foreground">
              Off by default. Only documents fetched over http(s):// or s3:// can be auto-refreshed.
            </p>
          </div>
          <Switch id="auto-refresh-toggle" checked={enabled} onCheckedChange={setEnabled} />
        </div>

        {enabled && (
          <div className="space-y-1.5">
            <Label htmlFor="auto-refresh-interval" className="text-xs text-muted-foreground">
              Check every (hours)
            </Label>
            <Input
              id="auto-refresh-interval"
              type="number"
              min={MIN_INTERVAL_HOURS}
              step={1}
              value={hours}
              onChange={e => setHours(e.target.value)}
              className="max-w-[100px]"
            />
            <p className="text-xs text-muted-foreground">
              At least {MIN_INTERVAL_HOURS} hour.
              {Number(hours) >= 24 ? ` (${describeInterval(Number(hours) * 3600)})` : ""}
            </p>
          </div>
        )}

        <div className="flex justify-end border-t border-border pt-4">
          <Button size="sm" onClick={handleSave} disabled={saving || !dirty}>
            {saving ? "Saving…" : "Save"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
