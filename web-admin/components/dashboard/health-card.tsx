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

const chipColors: Record<HealthStatus, string> = {
  loading: "bg-muted text-muted-foreground",
  ok: "bg-status-success-bg text-status-success-text",
  error: "bg-status-error-bg text-status-error-text",
};

const accentColors: Record<HealthStatus, string> = {
  loading: "border-l-transparent",
  ok: "border-l-status-success-border",
  error: "border-l-status-error-border",
};

export function HealthCard({ label, status, model, detail, icon = "embedder" }: HealthCardProps) {
  const Icon = iconMap[icon];

  return (
    <Card className={cn("border-l-4 transition-colors", accentColors[status])}>
      <CardHeader className="pb-2">
        <div className="flex items-center gap-3">
          <div className={cn("w-8 h-8 rounded-md flex items-center justify-center shrink-0 transition-colors", chipColors[status])}>
            <Icon className="w-4 h-4" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between gap-2">
              <CardTitle className="text-sm">{label}</CardTitle>
              <StatusIcon status={status} />
            </div>
            {model && (
              <CardDescription className="font-mono text-xs mt-0.5 truncate">{model}</CardDescription>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <p className={cn(
          "text-xs",
          status === "loading" && "text-muted-foreground",
          status === "ok" && "text-status-success-text",
          status === "error" && "text-status-error-text",
        )}>
          {status === "loading" ? "Checking…" : status === "ok" ? "Reachable" : detail ?? "Not reachable"}
        </p>
      </CardContent>
    </Card>
  );
}

function StatusIcon({ status }: { status: HealthStatus }) {
  if (status === "loading") {
    return <span className="w-2 h-2 rounded-full bg-muted-foreground/40 animate-pulse" />;
  }
  if (status === "ok") {
    return <CheckCircle className="w-4 h-4 text-status-success-text" />;
  }
  return <XCircle className="w-4 h-4 text-status-error-text" />;
}
