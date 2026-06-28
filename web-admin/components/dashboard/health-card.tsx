import { CheckCircle, XCircle, Cpu, Sparkles, Server } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type HealthStatus = "loading" | "ok" | "error";

interface HealthCardProps {
  label: string;
  status: HealthStatus;
  model?: string;
  detail?: string;
  icon?: "backend" | "embedder" | "llm";
}

const iconMap = {
  backend: Server,
  embedder: Cpu,
  llm: Sparkles,
};

export function HealthCard({ label, status, model, detail, icon = "embedder" }: HealthCardProps) {
  const Icon = iconMap[icon];

  return (
    <Card className={cn(
      "transition-colors",
      status === "ok" && "ring-emerald-200",
      status === "error" && "ring-red-200",
    )}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Icon className="w-4 h-4 text-muted-foreground" />
            <CardTitle className="text-sm">{label}</CardTitle>
          </div>
          <StatusIcon status={status} />
        </div>
        {model && (
          <CardDescription className="font-mono text-xs mt-1">{model}</CardDescription>
        )}
      </CardHeader>
      <CardContent className="pt-0">
        <p className={cn(
          "text-xs",
          status === "loading" && "text-muted-foreground",
          status === "ok" && "text-emerald-600",
          status === "error" && "text-red-500",
        )}>
          {status === "loading" ? "Checking…" : status === "ok" ? "Reachable" : detail ?? "Not reachable"}
        </p>
      </CardContent>
    </Card>
  );
}

function StatusIcon({ status }: { status: HealthStatus }) {
  if (status === "loading") {
    return <span className="w-2 h-2 rounded-full bg-zinc-300 animate-pulse" />;
  }
  if (status === "ok") {
    return <CheckCircle className="w-4 h-4 text-emerald-500" />;
  }
  return <XCircle className="w-4 h-4 text-red-400" />;
}
