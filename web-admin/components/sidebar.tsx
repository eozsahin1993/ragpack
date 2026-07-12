"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Logo } from "@/components/logo";
import { LayoutDashboard, Database, FileText, FlaskConical, ScrollText, KeyRound } from "lucide-react";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard, exact: true },
  { href: "/collections", label: "Collections", icon: Database },
  { href: "/documents", label: "Documents", icon: FileText },
  { href: "/prompts", label: "Prompts", icon: ScrollText },
  { href: "/playground", label: "Playground", icon: FlaskConical },
  { href: "/keys", label: "API Keys", icon: KeyRound },
];

export function Sidebar() {
  const path = usePathname();

  return (
    <aside className="w-56 shrink-0 flex flex-col bg-card border-r border-border h-screen">
      <div className="px-5 py-5 border-b border-border shrink-0">
        <Logo height={28} />
      </div>
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {nav.map(({ href, label, icon: Icon, exact }) => {
          const active = exact ? path === href : path === href || path.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors",
                active
                  ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
                  : "text-muted-foreground hover:text-foreground hover:bg-accent"
              )}
            >
              <Icon className="w-4 h-4 shrink-0" />
              {label}
            </Link>
          );
        })}
      </nav>
      <div className="px-5 py-4 border-t border-border">
        <p className="text-muted-foreground text-xs">v0.1</p>
      </div>
    </aside>
  );
}
