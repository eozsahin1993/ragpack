"use client";

import { KeyRound, Trash2, ShieldCheck } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { TableCell, TableRow } from "@/components/ui/table";
import { DataTable } from "@/components/data-table";
import { ApiKey, Collection, Permission, AdminResourceType } from "@/lib/api";

const PERMISSION_LABEL: Record<Permission, string> = {
  read: "Read",
  write: "Write",
  both: "Read/Write",
};

const RESOURCE_LABEL: Record<AdminResourceType, string> = {
  keys: "API Keys",
  prompts: "Prompts",
  collections: "Collections",
  "*": "Everything",
};

interface KeysTableProps {
  keys: ApiKey[];
  collections: Collection[];
  loading: boolean;
  onDelete: (key: ApiKey) => void;
}

export function KeysTable({ keys, collections, loading, onDelete }: KeysTableProps) {
  const collectionsById = new Map(collections.map(c => [c.id, c]));

  function collectionLabel(collectionId?: string) {
    if (!collectionId) return "All collections";
    return collectionsById.get(collectionId)?.name ?? "Unknown collection";
  }

  function renderCapabilities(k: ApiKey) {
    const grants = k.grants ?? [];
    const adminGrants = k.admin_grants ?? [];
    if (grants.length === 0 && adminGrants.length === 0) {
      return <span className="text-muted-foreground text-xs">No access</span>;
    }
    return (
      <div className="flex flex-wrap gap-1.5">
        {grants.map(g => (
          <Badge key={g.id} variant={g.collection_id ? "secondary" : "outline"} className="font-normal">
            {collectionLabel(g.collection_id)} · {PERMISSION_LABEL[g.permission]}
          </Badge>
        ))}
        {adminGrants.map(g => (
          <Badge key={g.id} variant="secondary" className="font-normal gap-1">
            <ShieldCheck className="w-3 h-3" />
            {RESOURCE_LABEL[g.resource_type]} (admin) · {PERMISSION_LABEL[g.permission]}
          </Badge>
        ))}
      </div>
    );
  }

  return (
    <DataTable columns={[
      { label: "Name" },
      { label: "Key" },
      { label: "Capabilities" },
      { label: "", className: "w-10" },
    ]}>
      {loading ? (
        <TableRow key="loading">
          <TableCell colSpan={4} className="text-center text-muted-foreground py-8">Loading…</TableCell>
        </TableRow>
      ) : keys.length === 0 ? (
        <TableRow key="empty">
          <TableCell colSpan={4} className="text-center text-muted-foreground py-8">
            No API keys yet. Create one to authenticate external clients.
          </TableCell>
        </TableRow>
      ) : keys.map(k => (
        <TableRow key={k.id} className="group">
          <TableCell>
            <div className="flex items-center gap-2">
              <KeyRound className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
              <span className="font-medium">{k.name}</span>
            </div>
          </TableCell>
          <TableCell>
            <code className="text-xs text-muted-foreground">rp_••••{k.key_hint}</code>
          </TableCell>
          <TableCell className="max-w-xl">{renderCapabilities(k)}</TableCell>
          <TableCell>
            <button
              onClick={() => onDelete(k)}
              className="text-muted-foreground/50 hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </TableCell>
        </TableRow>
      ))}
    </DataTable>
  );
}
