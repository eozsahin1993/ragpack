"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Database, BriefcaseBusiness, FlaskConical, ScrollText } from "lucide-react";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/collections", label: "Collections", icon: Database },
  { href: "/jobs", label: "Jobs", icon: BriefcaseBusiness },
  { href: "/prompts", label: "Prompts", icon: ScrollText },
  { href: "/playground", label: "Playground", icon: FlaskConical },
];

export function Sidebar() {
  const path = usePathname();

  return (
    <aside className="w-56 shrink-0 flex flex-col bg-zinc-900 border-r border-zinc-800 h-screen">
      <div className="px-5 py-5 border-b border-zinc-800">
        <span className="text-white font-semibold text-sm tracking-tight">ragpack</span>
      </div>
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = path === href || path.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors",
                active
                  ? "bg-zinc-800 text-white"
                  : "text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60"
              )}
            >
              <Icon className="w-4 h-4 shrink-0" />
              {label}
            </Link>
          );
        })}
      </nav>
      <div className="px-5 py-4 border-t border-zinc-800">
        <p className="text-zinc-600 text-xs">v0.1</p>
      </div>
    </aside>
  );
}
