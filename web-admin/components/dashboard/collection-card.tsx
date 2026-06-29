"use client";

import { useRouter } from "next/navigation";
import { FileText, ChevronRight } from "lucide-react";
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
      className="cursor-pointer hover:ring-zinc-300 transition-all group"
      onClick={() => router.push(`/collections/${collection.slug}`)}
    >
      <CardHeader>
        <div className="flex items-start justify-between">
          <CardTitle className="truncate">{collection.name}</CardTitle>
          <ChevronRight className="w-4 h-4 text-muted-foreground shrink-0 mt-0.5 group-hover:translate-x-0.5 transition-transform" />
        </div>
        <CardDescription className="font-mono text-xs">{collection.embed_model}</CardDescription>
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
