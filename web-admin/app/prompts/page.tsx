"use client";

import { useEffect, useState } from "react";
import { Plus, Trash2, Pencil } from "lucide-react";
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api, Prompt } from "@/lib/api";

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState({ name: "", content: "" });
  const [creating, setCreating] = useState(false);

  const [editTarget, setEditTarget] = useState<Prompt | null>(null);
  const [editForm, setEditForm] = useState({ name: "", content: "" });
  const [saving, setSaving] = useState(false);

  async function load() {
    try {
      const data = await api.prompts.list();
      setPrompts(data.prompts ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError("");
    try {
      await api.prompts.create({ name: createForm.name, content: createForm.content });
      setCreateForm({ name: "", content: "" });
      setCreateOpen(false);
      await load();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to create");
    } finally {
      setCreating(false);
    }
  }

  function openEdit(p: Prompt) {
    setEditTarget(p);
    setEditForm({ name: p.name, content: p.content });
  }

  async function handleEdit(e: React.FormEvent) {
    e.preventDefault();
    if (!editTarget) return;
    setSaving(true);
    setError("");
    try {
      await api.prompts.update(editTarget.slug, {
        name: editForm.name !== editTarget.name ? editForm.name : undefined,
        content: editForm.content !== editTarget.content ? editForm.content : undefined,
      });
      setEditTarget(null);
      await load();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  function renderRows() {
    if (loading) return (
      <TableRow>
        <TableCell colSpan={4} className="text-center text-zinc-400 py-10">Loading…</TableCell>
      </TableRow>
    );
    if (prompts.length === 0) return (
      <TableRow>
        <TableCell colSpan={4} className="text-center text-zinc-400 py-10">
          No prompts yet. Create one to get started.
        </TableCell>
      </TableRow>
    );
    return prompts.map(p => (
      <TableRow key={p.id} className="group">
        <TableCell className="font-medium">{p.name}</TableCell>
        <TableCell>
          <Badge variant="secondary" className="font-mono text-xs">{p.slug}</Badge>
        </TableCell>
        <TableCell className="text-zinc-500 text-sm max-w-sm truncate">
          {p.content}
        </TableCell>
        <TableCell>
          <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button onClick={() => openEdit(p)} className="text-zinc-400 hover:text-zinc-700">
              <Pencil className="w-4 h-4" />
            </button>
            <button onClick={() => handleDelete(p.slug, p.name)} className="text-zinc-400 hover:text-red-500">
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </TableCell>
      </TableRow>
    ));
  }

  async function handleDelete(slug: string, name: string) {
    if (!confirm(`Delete "${name}"?`)) return;
    try {
      await api.prompts.delete(slug);
      await load();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to delete");
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Prompts</h1>
          <p className="text-sm text-zinc-500 mt-0.5">System prompts for RAG generation</p>
        </div>
        <Button size="sm" className="gap-2" onClick={() => setCreateOpen(true)}>
          <Plus className="w-4 h-4" /> New prompt
        </Button>

        {/* Create dialog */}
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>New prompt</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4 pt-2">
              <div className="space-y-1.5">
                <Label>Name</Label>
                <Input
                  required
                  value={createForm.name}
                  onChange={e => setCreateForm({ ...createForm, name: e.target.value })}
                  placeholder="Customer support"
                />
              </div>
              <div className="space-y-1.5">
                <Label>Content</Label>
                <textarea
                  required
                  rows={6}
                  value={createForm.content}
                  onChange={e => setCreateForm({ ...createForm, content: e.target.value })}
                  placeholder="You are a helpful assistant. Answer using only the provided context..."
                  className="w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 resize-none"
                />
              </div>
              {error && <p className="text-red-500 text-sm">{error}</p>}
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={creating}>{creating ? "Creating…" : "Create"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>

        {/* Edit dialog */}
        <Dialog open={!!editTarget} onOpenChange={open => !open && setEditTarget(null)}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>Edit prompt</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleEdit} className="space-y-4 pt-2">
              <div className="space-y-1.5">
                <Label>Name</Label>
                <Input
                  required
                  value={editForm.name}
                  onChange={e => setEditForm({ ...editForm, name: e.target.value })}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Content</Label>
                <textarea
                  required
                  rows={6}
                  value={editForm.content}
                  onChange={e => setEditForm({ ...editForm, content: e.target.value })}
                  className="w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 resize-none"
                />
              </div>
              {error && <p className="text-red-500 text-sm">{error}</p>}
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => setEditTarget(null)}>Cancel</Button>
                <Button type="submit" disabled={saving}>{saving ? "Saving…" : "Save"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {error && !createOpen && !editTarget && <p className="text-red-500 text-sm">{error}</p>}

      <div className="rounded-lg border bg-white overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="bg-zinc-50">
              <TableHead>Name</TableHead>
              <TableHead>Slug</TableHead>
              <TableHead>Preview</TableHead>
              <TableHead className="w-20"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {renderRows()}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
