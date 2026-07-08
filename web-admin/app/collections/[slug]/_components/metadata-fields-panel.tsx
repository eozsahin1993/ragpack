"use client";

import { useEffect, useState, useCallback } from "react";
import { Plus, RefreshCw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { api, MetadataField } from "@/lib/api";

const TYPE_LABELS: Record<MetadataField["type"], string> = {
  str: "String",
  num: "Number",
  bool: "Boolean",
  date: "Date",
  arr: "String array",
};

const TYPE_COLORS: Record<MetadataField["type"], string> = {
  str: "bg-blue-50 text-blue-700 border-blue-200",
  num: "bg-amber-50 text-amber-700 border-amber-200",
  bool: "bg-purple-50 text-purple-700 border-purple-200",
  date: "bg-green-50 text-green-700 border-green-200",
  arr: "bg-zinc-100 text-zinc-600 border-zinc-200",
};

interface Props {
  slug: string;
}

export function MetadataFieldsPanel({ slug }: Props) {
  const [fields, setFields] = useState<MetadataField[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState<MetadataField["type"]>("str");
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  async function load() {
    try {
      const data = await api.metadataFields.list(slug);
      setFields(data.fields ?? []);
    } catch {
      // non-fatal
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, [slug]);

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    setSaving(true);
    try {
      const data = await api.metadataFields.register(slug, [{ name: name.trim(), type }]);
      setFields(prev => [...prev, ...(data.fields ?? [])]);
      setName("");
      setType("str");
      setOpen(false);
      toast.success(`Field "${name.trim()}" registered`);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to register field");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(fieldName: string) {
    if (!confirm(`Delete field "${fieldName}"? Existing data in this slot will be nulled out.`)) return;
    setDeleting(fieldName);
    try {
      await api.metadataFields.delete(slug, fieldName);
      setFields(prev => prev.filter(f => f.name !== fieldName));
      toast.success(`Field "${fieldName}" deleted`);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setDeleting(null);
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-zinc-500">
          {loading ? "Loading…" : `${fields.length} field${fields.length !== 1 ? "s" : ""} registered`}
        </p>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90">
            <Plus className="w-3.5 h-3.5" />
            Add field
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Register metadata field</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleRegister} className="space-y-4 pt-2">
              <div className="space-y-1.5">
                <Label className="text-xs text-zinc-500">Field name</Label>
                <Input
                  required
                  placeholder="e.g. author, publish_date, score"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  autoFocus
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-zinc-500">Type</Label>
                <Select value={type} onValueChange={v => setType(v as MetadataField["type"])}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {(Object.entries(TYPE_LABELS) as [MetadataField["type"], string][]).map(([val, label]) => (
                      <SelectItem key={val} value={val}>{label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex justify-end gap-2">
                <Button type="button" variant="ghost" onClick={() => setOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={saving}>{saving ? "Saving…" : "Register"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {!loading && fields.length === 0 ? (
        <div className="rounded-lg border bg-white px-6 py-10 text-center text-zinc-400 text-sm">
          No metadata fields yet. Add one to enable filtering and enriched results.
        </div>
      ) : (
        <div className="rounded-lg border bg-white divide-y">
          {fields.map(f => (
            <div key={f.id} className="flex items-center justify-between px-4 py-3">
              <div className="flex items-center gap-3">
                <span className="font-mono text-sm font-medium">{f.name}</span>
                <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${TYPE_COLORS[f.type]}`}>
                  {TYPE_LABELS[f.type]}
                </span>
                <span className="text-xs text-zinc-400">slot {f.slot}</span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="text-red-400 hover:text-red-600 hover:bg-red-50 h-7 w-7 p-0"
                onClick={() => handleDelete(f.name)}
                disabled={deleting === f.name}
              >
                <Trash2 className="w-3.5 h-3.5" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
