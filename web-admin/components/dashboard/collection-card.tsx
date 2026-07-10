"use client";

import { useRouter } from "next/navigation";
import { Database, FileText, ChevronRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Collection } from "@/lib/api";

interface CollectionCardProps {
  collection: Collection;
  docCount: number | null;
}

export function CollectionCard({ collection, docCount }: CollectionCardProps) {
  const router = useRouter();

  return (
    <Card
      className="cursor-pointer hover:shadow-md hover:ring-2 hover:ring-primary/20 transition-all group"
      onClick={() => router.push(`/collections/${collection.slug}`)}
    >
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-8 h-8 rounded-md flex items-center justify-center shrink-0 bg-accent text-primary">
              <Database className="w-4 h-4" />
            </div>
            <div className="min-w-0">
              <CardTitle className="truncate">{collection.name}</CardTitle>
              <CardDescription className="font-mono text-xs mt-0.5">{collection.embed_model}</CardDescription>
            </div>
          </div>
          <ChevronRight className="w-4 h-4 text-muted-foreground shrink-0 mt-1 group-hover:translate-x-0.5 group-hover:text-primary transition-all" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4 text-sm">
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <FileText className="w-3.5 h-3.5" />
            <span>{docCount != null ? `${docCount} doc${docCount === 1 ? "" : "s"}` : "—"}</span>
          </div>
          <span className="text-muted-foreground text-xs">{collection.vector_dim}d</span>
        </div>
      </CardContent>
    </Card>
  );
}
