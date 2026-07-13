"use client";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const PRESETS = [
  { value: "7", label: "Last 7 days" },
  { value: "30", label: "Last 30 days" },
  { value: "90", label: "Last 90 days" },
];

interface DaysSelectProps {
  days: number;
  onChange: (days: number) => void;
}

// The one filter row above every chart (see dataviz skill's interaction.md) —
// every panel on the page scopes to this same value, so the numbers always
// agree with each other.
export function DaysSelect({ days, onChange }: DaysSelectProps) {
  return (
    <Select items={PRESETS} value={String(days)} onValueChange={v => v && onChange(Number(v))}>
      <SelectTrigger className="w-40">
        <SelectValue placeholder="Date range" />
      </SelectTrigger>
      <SelectContent>
        {PRESETS.map(p => (
          <SelectItem key={p.value} value={p.value}>{p.label}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
