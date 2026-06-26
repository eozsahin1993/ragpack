"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const tabs = [
  { href: "/playground/search", label: "Search" },
  { href: "/playground/rag", label: "RAG" },
];

export default function PlaygroundLayout({ children }: { children: React.ReactNode }) {
  const path = usePathname();

  return (
    <div className="space-y-6 max-w-3xl">
      <div>
        <h1 className="text-xl font-semibold">Playground</h1>
        <p className="text-sm text-zinc-500 mt-0.5">Test your collections interactively</p>
      </div>

      <div className="flex gap-1 border-b border-zinc-200">
        {tabs.map(t => (
          <Link
            key={t.href}
            href={t.href}
            className={cn(
              "px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
              path === t.href
                ? "border-zinc-900 text-zinc-900"
                : "border-transparent text-zinc-500 hover:text-zinc-700"
            )}
          >
            {t.label}
          </Link>
        ))}
      </div>

      {children}
    </div>
  );
}
