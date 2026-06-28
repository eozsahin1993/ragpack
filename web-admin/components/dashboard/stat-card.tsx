import { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface StatCardProps {
  label: string;
  value: string | number;
  icon: LucideIcon;
  loading?: boolean;
  accent?: "default" | "amber" | "red";
  className?: string;
}

export function StatCard({ label, value, icon: Icon, loading, accent = "default", className }: StatCardProps) {
  const iconColors = {
    default: "bg-zinc-100 text-zinc-500",
    amber: "bg-amber-50 text-amber-500",
    red: "bg-red-50 text-red-400",
  };

  return (
    <Card className={cn("rounded-xl", className)}>
      <CardContent className="pt-4 flex items-center gap-4">
        <div className={cn("w-9 h-9 rounded-md flex items-center justify-center shrink-0", iconColors[accent])}>
          <Icon className="w-4 h-4" />
        </div>
        <div>
          <p className="text-2xl font-semibold">{loading ? "—" : value}</p>
          <p className="text-xs text-muted-foreground mt-0.5">{label}</p>
        </div>
      </CardContent>
    </Card>
  );
}
