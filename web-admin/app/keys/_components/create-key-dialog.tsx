"use client";

import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import {
  api,
  Collection,
  Permission,
  AdminResourceType,
  CreateApiKeyGrant,
  CreateApiKeyAdminGrant,
  CreatedApiKey,
} from "@/lib/api";

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

type GrantRow = { scope: "all" | string; permission: Permission };
type AdminGrantRow = { resourceType: AdminResourceType; permission: Permission };

function emptyGrantRow(): GrantRow {
  return { scope: "all", permission: "read" };
}

function emptyAdminRow(): AdminGrantRow {
  return { resourceType: "keys", permission: "read" };
}

// GrantRowEditor is shared by the collection-access and admin-access lists
// below — both are "a repeatable row with some scope selector + a
// permission selector + a remove button, plus an inline add-row control at
// the bottom." The only thing that actually differs between the two is
// what the scope selector looks like, so that's the one piece callers
// supply via renderFirstField.
interface GrantRowEditorProps<T> {
  rows: T[];
  onChange: (rows: T[]) => void;
  newRow: () => T;
  getPermission: (row: T) => Permission;
  setPermission: (row: T, permission: Permission) => T;
  renderFirstField: (row: T, onChange: (patch: Partial<T>) => void) => React.ReactNode;
  emptyMessage?: string;
}

function GrantRowEditor<T>({
  rows,
  onChange,
  newRow,
  getPermission,
  setPermission,
  renderFirstField,
  emptyMessage,
}: GrantRowEditorProps<T>) {
  function updateRow(i: number, updater: (row: T) => T) {
    onChange(rows.map((r, idx) => (idx === i ? updater(r) : r)));
  }

  return (
    <>
      {rows.length === 0 && emptyMessage && (
        <p className="text-xs text-muted-foreground">{emptyMessage}</p>
      )}
      <div className="space-y-2">
        {rows.map((row, i) => (
          <div key={i} className="flex items-center gap-2">
            {renderFirstField(row, patch => updateRow(i, r => ({ ...r, ...patch })))}
            <Select
              value={getPermission(row)}
              onValueChange={v => v && updateRow(i, r => setPermission(r, v as Permission))}
            >
              <SelectTrigger className="w-36">
                <span>{PERMISSION_LABEL[getPermission(row)]}</span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="read">Read</SelectItem>
                <SelectItem value="write">Write</SelectItem>
                <SelectItem value="both">Read/Write</SelectItem>
              </SelectContent>
            </Select>
            <button
              type="button"
              onClick={() => onChange(rows.filter((_, idx) => idx !== i))}
              className="text-muted-foreground/50 hover:text-destructive shrink-0"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={() => onChange([...rows, newRow()])}
          className="w-full flex items-center justify-center gap-1.5 rounded-lg border border-dashed border-input py-1.5 text-xs text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
        >
          <Plus className="w-3.5 h-3.5" /> Add grant
        </button>
      </div>
    </>
  );
}

interface CreateKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  collections: Collection[];
  onCreated: (key: CreatedApiKey) => void;
}

export function CreateKeyDialog({ open, onOpenChange, collections, onCreated }: CreateKeyDialogProps) {
  const [name, setName] = useState("");
  const [grantRows, setGrantRows] = useState<GrantRow[]>([]);
  const [adminRows, setAdminRows] = useState<AdminGrantRow[]>([]);
  const [creating, setCreating] = useState(false);

  const collectionsBySlug = new Map(collections.map(c => [c.slug, c]));

  function reset() {
    setName("");
    setGrantRows([]);
    setAdminRows([]);
  }

  // Every path that closes the dialog — backdrop/Escape, Cancel, or a
  // successful create — must go through this, not just the Dialog
  // primitive's own onOpenChange (which only fires for backdrop/Escape,
  // not for an external close call like the Cancel button's onClick).
  function close() {
    onOpenChange(false);
    reset();
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (grantRows.length === 0 && adminRows.length === 0) {
      toast.error("At least one grant — collection or admin — is required");
      return;
    }
    setCreating(true);
    try {
      const grants: CreateApiKeyGrant[] = grantRows.map(r => ({
        ...(r.scope !== "all" ? { collection_slug: r.scope } : {}),
        permission: r.permission,
      }));
      const admin_grants: CreateApiKeyAdminGrant[] = adminRows.map(r => ({
        resource_type: r.resourceType,
        permission: r.permission,
      }));
      const created = await api.keys.create({
        name,
        grants,
        admin_grants: admin_grants.length > 0 ? admin_grants : undefined,
      });
      close();
      onCreated(created);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to create API key");
    } finally {
      setCreating(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={o => (o ? onOpenChange(o) : close())}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>New API key</DialogTitle>
          <DialogDescription>
            Grant collection access, instance-admin access, or both. At least one grant of either kind is required.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleCreate} className="space-y-5 pt-2">
          <div className="space-y-1.5">
            <Label>Name</Label>
            <Input
              required
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="New API Key"
            />
          </div>

          <div className="space-y-2">
            <Label>Collection access</Label>
            <GrantRowEditor<GrantRow>
              rows={grantRows}
              onChange={setGrantRows}
              newRow={emptyGrantRow}
              getPermission={r => r.permission}
              setPermission={(r, permission) => ({ ...r, permission })}
              emptyMessage="No collection access — this will be an admin-only key."
              renderFirstField={(row, onChange) => (
                <Select value={row.scope} onValueChange={v => v && onChange({ scope: v })}>
                  <SelectTrigger className="flex-1">
                    <span>{row.scope === "all" ? "All collections" : collectionsBySlug.get(row.scope)?.name ?? row.scope}</span>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All collections</SelectItem>
                    {collections.map(c => (
                      <SelectItem key={c.id} value={c.slug}>{c.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </div>

          <div className="space-y-2">
            <div>
              <Label>Admin access</Label>
              <p className="text-xs text-muted-foreground mt-0.5">
                Manage API keys, prompts, or collections themselves — independent of collection content access above.
              </p>
            </div>
            <GrantRowEditor<AdminGrantRow>
              rows={adminRows}
              onChange={setAdminRows}
              newRow={emptyAdminRow}
              getPermission={r => r.permission}
              setPermission={(r, permission) => ({ ...r, permission })}
              renderFirstField={(row, onChange) => (
                <Select value={row.resourceType} onValueChange={v => v && onChange({ resourceType: v as AdminResourceType })}>
                  <SelectTrigger className="flex-1">
                    <span>{RESOURCE_LABEL[row.resourceType]}</span>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="keys">API Keys</SelectItem>
                    <SelectItem value="prompts">Prompts</SelectItem>
                    <SelectItem value="collections">Collections</SelectItem>
                    <SelectItem value="*">Everything</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={close}>Cancel</Button>
            <Button type="submit" disabled={creating}>{creating ? "Creating…" : "Create key"}</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
