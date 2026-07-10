"use client";

import { useEffect, useState } from "react";
import { Plus, Trash2, Pencil, Lock, ChevronDown, ChevronRight } from "lucide-react";
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
  TableCell,
  TableRow,
} from "@/components/ui/table";
import { api, Prompt } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { DataTable } from "@/components/data-table";

export default function PromptsPage() {
  const [systemPrompts, setSystemPrompts] = useState<Prompt[]>([]);
  const [userPrompts, setUserPrompts] = useState<Prompt[]>([]);
  const [loading, setLoading] = useState(true);

  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState({ name: "", content: "" });
  const [creating, setCreating] = useState(false);

  const [editTarget, setEditTarget] = useState<Prompt | null>(null);
  const [editForm, setEditForm] = useState({ name: "", content: "" });
  const [saving, setSaving] = useState(false);
  const [expandedSlug, setExpandedSlug] = useState<string | null>(null);

  async function load() {
    try {
      const data = await api.prompts.list();
      setSystemPrompts(data.system ?? []);
      setUserPrompts(data.user ?? []);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to load prompts");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    try {
      await api.prompts.create({ name: createForm.name, content: createForm.content });
      setCreateForm({ name: "", content: "" });
      setCreateOpen(false);
      await load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to create prompt");
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
    try {
      await api.prompts.update(editTarget.slug, {
        name: editForm.name !== editTarget.name ? editForm.name : undefined,
        content: editForm.content !== editTarget.content ? editForm.content : undefined,
      });
      setEditTarget(null);
      await load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to save prompt");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(slug: string, name: string) {
    if (!confirm(`Delete "${name}"?`)) return;
    try {
      await api.prompts.delete(slug);
      await load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to delete prompt");
    }
  }

  function renderPlaceholders(content: string) {
    return content.split(/(\{\{[^}]+\}\})/g).map((part, i) =>
      /^\{\{[^}]+\}\}$/.test(part)
        ? <mark key={i} className="bg-accent text-primary rounded px-0.5 not-italic font-medium">{part}</mark>
        : <span key={i}>{part}</span>
    );
  }

  function renderPromptRows(rows: Prompt[], isSystem: boolean) {
    return rows.flatMap((p, i) => {
      const isExpanded = expandedSlug === (p.slug || String(i));
      const key = p.id || p.slug || String(i);
      return [
        <TableRow key={key} className="group cursor-pointer" onClick={() => setExpandedSlug(isExpanded ? null : (p.slug || String(i)))}>
          <TableCell>
            <div className="flex items-center gap-2">
              {isExpanded
                ? <ChevronDown className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                : <ChevronRight className="w-3.5 h-3.5 text-muted-foreground shrink-0" />}
              <span className="font-medium">{p.name}</span>
            </div>
          </TableCell>
          <TableCell>
            <Badge variant="secondary" className="font-mono text-xs">{p.slug}</Badge>
          </TableCell>
          <TableCell className="text-muted-foreground text-sm max-w-sm truncate">
            {p.content}
          </TableCell>
          <TableCell onClick={e => e.stopPropagation()}>
            {isSystem ? (
              <Lock className="w-3.5 h-3.5 text-muted-foreground/50" />
            ) : (
              <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <button onClick={() => openEdit(p)} className="text-muted-foreground/50 hover:text-primary">
                  <Pencil className="w-4 h-4" />
                </button>
                <button onClick={() => handleDelete(p.slug, p.name)} className="text-muted-foreground/50 hover:text-destructive">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            )}
          </TableCell>
        </TableRow>,
        isExpanded && (
          <TableRow key={`${key}-expanded`}>
            <TableCell colSpan={4} className="bg-muted/40 px-6 pb-5 pt-3">
              <pre className="text-xs text-foreground font-mono whitespace-pre-wrap leading-relaxed">
                {renderPlaceholders(p.content)}
              </pre>
            </TableCell>
          </TableRow>
        ),
      ].filter(Boolean);
    });
  }

  return (
    <div className="space-y-8">
      <PageHeader
        title="Prompts"
        description={<>RAG prompt templates using <code className="text-xs bg-muted px-1 py-0.5 rounded">{"{{context}}"}</code> and <code className="text-xs bg-muted px-1 py-0.5 rounded">{"{{question}}"}</code></>}
        action={<Button className="gap-2" onClick={() => setCreateOpen(true)}><Plus className="w-4 h-4" /> New prompt</Button>}
      />

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
                  rows={8}
                  value={createForm.content}
                  onChange={e => setCreateForm({ ...createForm, content: e.target.value })}
                  placeholder={"You are a helpful assistant. Answer using only the provided context.\n\nContext:\n{{context}}\n\nQuestion: {{question}}\n\nAnswer:"}
                  className="w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm font-mono outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 resize-none"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={creating}>{creating ? "Creating…" : "Create"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>

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
                  rows={8}
                  value={editForm.content}
                  onChange={e => setEditForm({ ...editForm, content: e.target.value })}
                  className="w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm font-mono outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 resize-none"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button type="button" variant="outline" onClick={() => setEditTarget(null)}>Cancel</Button>
                <Button type="submit" disabled={saving}>{saving ? "Saving…" : "Save"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>

      {/* System prompts */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <h2 className="text-sm font-medium text-foreground">Built-in</h2>
          <Lock className="w-3 h-3 text-muted-foreground" />
        </div>
        <DataTable columns={[
          { label: "Name" },
          { label: "Slug" },
          { label: "Preview" },
          { label: "", className: "w-10" },
        ]}>
          {loading ? (
            <TableRow key="loading">
              <TableCell colSpan={4} className="text-center text-muted-foreground py-8">Loading…</TableCell>
            </TableRow>
          ) : systemPrompts.length === 0 ? (
            <TableRow key="empty">
              <TableCell colSpan={4} className="text-center text-muted-foreground py-8">No built-in prompts.</TableCell>
            </TableRow>
          ) : renderPromptRows(systemPrompts, true)}
        </DataTable>
      </div>

      {/* User prompts */}
      <div className="space-y-2">
        <h2 className="text-sm font-medium text-foreground">Custom</h2>
        <DataTable columns={[
          { label: "Name" },
          { label: "Slug" },
          { label: "Preview" },
          { label: "", className: "w-20" },
        ]}>
          {loading ? (
            <TableRow key="loading">
              <TableCell colSpan={4} className="text-center text-muted-foreground py-8">Loading…</TableCell>
            </TableRow>
          ) : userPrompts.length === 0 ? (
            <TableRow key="empty">
              <TableCell colSpan={4} className="text-center text-muted-foreground py-8">
                No custom prompts yet. Create one to get started.
              </TableCell>
            </TableRow>
          ) : renderPromptRows(userPrompts, false)}
        </DataTable>
      </div>
    </div>
  );
}
