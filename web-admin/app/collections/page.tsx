"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Trash2, ChevronRight, Database } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import {
  TableCell,
  TableRow,
} from "@/components/ui/table";
import { api, Collection, EmbedderInfo } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { DataTable } from "@/components/data-table";

export default function CollectionsPage() {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState({ name: "", embed_model: "" });
  const [creating, setCreating] = useState(false);
  const [embedderInfo, setEmbedderInfo] = useState<EmbedderInfo | null>(null);

  async function load() {
    try {
      const data = await api.collections.list();
      setCollections(data.collections ?? []);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to load collections");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    api.embedders.list().then(setEmbedderInfo).catch(() => {});
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    try {
      await api.collections.create({
        name: form.name,
        ...(form.embed_model.trim() && { embed_model: form.embed_model.trim() }),
      });
      setForm({ name: "", embed_model: "" });
      setOpen(false);
      await load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to create collection");
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(slug: string, name: string) {
    if (!confirm(`Delete "${name}"? This removes all indexed data.`)) return;
    try {
      await api.collections.delete(slug);
      await load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to delete collection");
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Collections"
        description="Manage your vector collections and indexed documents"
        action={<Button className="gap-2" onClick={() => setOpen(true)}><Plus className="w-4 h-4" /> New collection</Button>}
      />
      <Dialog open={open} onOpenChange={setOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>New collection</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4 pt-2">
              <div className="space-y-1.5">
                <Label>Name</Label>
                <Input
                  required
                  value={form.name}
                  onChange={e => setForm({ ...form, name: e.target.value })}
                  placeholder="My documents"
                />
              </div>
              {embedderInfo && (
                <div className="space-y-1.5">
                  <Label>
                    Embedding model{" "}
                    <span className="text-muted-foreground font-normal">(optional)</span>
                  </Label>
                  <Select
                    value={form.embed_model || "__default__"}
                    onValueChange={v => setForm({ ...form, embed_model: !v || v === "__default__" ? "" : v })}
                  >
                    <SelectTrigger className="w-full">
                      <span>
                        {form.embed_model || `Default (${embedderInfo.default})`}
                      </span>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__default__">
                        Default ({embedderInfo.default})
                      </SelectItem>
                      {embedderInfo.models
                        .filter(m => m !== embedderInfo.default)
                        .map(m => (
                          <SelectItem key={m} value={m}>{m}</SelectItem>
                        ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={creating}>
                  {creating ? "Creating…" : "Create"}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>

      <DataTable columns={[
        { label: "Name" },
        { label: "Slug" },
        { label: "Model" },
        { label: "Dimensions" },
        { label: "", className: "w-20" },
      ]}>
        {loading ? (
          <TableRow>
            <TableCell colSpan={5} className="text-center text-muted-foreground py-10">Loading…</TableCell>
          </TableRow>
        ) : collections.length === 0 ? (
          <TableRow>
            <TableCell colSpan={5} className="text-center text-muted-foreground py-10">
              No collections yet. Create one to get started.
            </TableCell>
          </TableRow>
        ) : collections.map(c => (
          <TableRow key={c.id} className="group">
            <TableCell>
              <Link href={`/collections/${c.slug}`} className="font-medium hover:text-primary flex items-center gap-2">
                <div className="w-7 h-7 rounded-md bg-accent text-primary flex items-center justify-center shrink-0">
                  <Database className="w-3.5 h-3.5" />
                </div>
                {c.name}
                <ChevronRight className="w-3.5 h-3.5 opacity-0 group-hover:opacity-50 transition-opacity" />
              </Link>
            </TableCell>
            <TableCell>
              <Badge variant="secondary" className="font-mono text-xs">{c.slug}</Badge>
            </TableCell>
            <TableCell className="text-muted-foreground text-sm">{c.embed_model}</TableCell>
            <TableCell className="text-muted-foreground text-sm">{c.vector_dim}</TableCell>
            <TableCell>
              <button
                onClick={() => handleDelete(c.slug, c.name)}
                className="opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground/50 hover:text-destructive"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </TableCell>
          </TableRow>
        ))}
      </DataTable>
    </div>
  );
}
