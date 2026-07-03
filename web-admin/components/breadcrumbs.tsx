"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight } from "lucide-react";

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
  const segments = pathname.split("/").filter(Boolean);

  if (segments.length <= 1) return null;

  const crumbs = segments.map((seg, i) => ({
    label: labelFor(seg),
    href: "/" + segments.slice(0, i + 1).join("/"),
  }));

  return (
    <nav className="flex items-center gap-1.5 text-sm mb-5">
      {crumbs.map((crumb, i) => (
        <span key={crumb.href} className="flex items-center gap-1.5">
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
