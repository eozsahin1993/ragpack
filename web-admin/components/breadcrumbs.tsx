"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight } from "lucide-react";
import { useBreadcrumbLabels } from "@/components/breadcrumb-context";

const LABELS: Record<string, string> = {
  collections: "Collections",
  playground:  "Playground",
  prompts:     "Prompts",
  jobs:        "Jobs",
  search:      "Search",
  rag:         "RAG",
  documents:   "Documents",
  chunks:      "Chunks",
};

const NO_BREADCRUMB = new Set(["playground"]);

function isUUID(s: string) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(s);
}

function labelFor(segment: string): string {
  if (LABELS[segment]) return LABELS[segment];
  if (isUUID(segment)) return segment.slice(0, 8) + "…";
  return segment.replace(/[_-]/g, " ").replace(/\b\w/g, c => c.toUpperCase());
}

export function Breadcrumbs() {
  const pathname = usePathname();
  const overrides = useBreadcrumbLabels();
  const segments = pathname.split("/").filter(Boolean);

  if (segments.length <= 1 || NO_BREADCRUMB.has(segments[0])) return null;

  // /collections/:slug/documents/:id/chunks has no page at .../documents or
  // .../documents/:id — only the full .../chunks path is real. "Documents"
  // there is just the collection page's default tab, so point it there
  // instead of a 404; point the doc-name crumb at the real chunks path.
  const isNestedDocRoute = segments[0] === "collections" && segments[2] === "documents";

  const crumbs = segments.map((seg, i) => {
    let href = "/" + segments.slice(0, i + 1).join("/");
    if (isNestedDocRoute && i === 2) href = "/" + segments.slice(0, 2).join("/");
    if (isNestedDocRoute && i === 3) href = "/" + segments.slice(0, 5).join("/");
    return { label: overrides[seg] ?? labelFor(seg), href };
  });

  return (
    <nav className="flex items-center gap-1.5 text-sm mb-5">
      {crumbs.map((crumb, i) => (
        <span key={i} className="flex items-center gap-1.5">
          {i > 0 && <ChevronRight className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />}
          {i === crumbs.length - 1 ? (
            <span className="font-medium text-foreground">{crumb.label}</span>
          ) : (
            <Link href={crumb.href} className="text-muted-foreground hover:text-foreground transition-colors">
              {crumb.label}
            </Link>
          )}
        </span>
      ))}
    </nav>
  );
}
