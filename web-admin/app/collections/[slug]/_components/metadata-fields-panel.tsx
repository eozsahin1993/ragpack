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
import { Card, CardContent } from "@/components/ui/card";
import { api, MetadataField } from "@/lib/api";

const TYPE_LABELS: Record<MetadataField["type"], string> = {
  str: "String",
  num: "Number",
  bool: "Boolean",
  date: "Date",
  arr: "String array",
};

// Distinct hues per type (categorical, not status) so the 5 types stay
// visually distinguishable at a glance — not a fit for the 3-color
// success/warning/error status tokens.
const TYPE_COLORS: Record<MetadataField["type"], string> = {
  str: "bg-blue-50 text-blue-700 border-blue-200",
  num: "bg-amber-50 text-amber-700 border-amber-200",
  bool: "bg-purple-50 text-purple-700 border-purple-200",
  date: "bg-green-50 text-green-700 border-green-200",
  arr: "bg-muted text-muted-foreground border-border",
};

interface Props {
  slug: string;
}

export function MetadataFieldsPanel({ slug }: Props) {
  const [fields, setFields] = useState<MetadataField[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState<MetadataField["type"]>("str");
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const data = await api.metadataFields.list(slug);
      setFields(data.fields ?? []);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to load metadata fields";
      setLoadError(msg);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  }, [slug]);

  useEffect(() => { load(); }, [load]);

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
        <div className="flex items-center gap-2">
          <p className="text-sm text-muted-foreground">
            {loading ? "Loading…" : loadError ? "Failed to load" : `${fields.length} propert${fields.length !== 1 ? "ies" : "y"} defined`}
          </p>
          {!loading && (
            <button
              type="button"
              onClick={load}
              className="text-muted-foreground/60 hover:text-primary transition-colors"
              title="Reload"
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90">
            <Plus className="w-3.5 h-3.5" />
            Add property
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add document property</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleRegister} className="space-y-4 pt-2">
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">Property name</Label>
                <Input
                  required
                  placeholder="e.g. author, publish_date, score"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  autoFocus
                />
              </div>
              <div className="space-y-1.5">
                <Label className="text-xs text-muted-foreground">Type</Label>
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
                <Button type="submit" disabled={saving}>{saving ? "Saving…" : "Add"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {!loading && fields.length === 0 ? (
        <Card>
          <CardContent className="px-6 py-10 text-center text-muted-foreground text-sm">
            No properties defined yet. Add one to enable filtering and enriched results.
          </CardContent>
        </Card>
      ) : (
        <Card className="py-0">
          <CardContent className="p-0 divide-y divide-border">
            {fields.map(f => (
              <div key={f.id} className="flex items-center justify-between px-4 py-3">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm font-medium">{f.name}</span>
                  <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${TYPE_COLORS[f.type]}`}>
                    {TYPE_LABELS[f.type]}
                  </span>
                  <span className="text-xs text-muted-foreground">slot {f.slot}</span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-destructive/70 hover:text-destructive hover:bg-destructive/10 h-7 w-7 p-0"
                  onClick={() => handleDelete(f.name)}
                  disabled={deleting === f.name}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
