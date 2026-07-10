"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { api, Document, MetadataField } from "@/lib/api";
import { MetaFieldInput } from "@/components/meta-field-input";

interface Props {
  slug: string;
  doc: Document | null;
  metadataFields: MetadataField[];
  onClose: () => void;
  onSaved: (doc: Document) => void;
}

function emptyMetadata(fields: MetadataField[]): Record<string, string> {
  return Object.fromEntries(fields.map(f => [f.name, ""]));
}

export function DocumentEditDialog({ slug, doc, metadataFields, onClose, onSaved }: Props) {
  const [name, setName] = useState("");
  const [metaValues, setMetaValues] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!doc) return;
    setName(doc.name ?? "");
    setMetaValues(emptyMetadata(metadataFields));
  }, [doc, metadataFields]);

  function setField(fieldName: string, value: string) {
    setMetaValues(prev => ({ ...prev, [fieldName]: value }));
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (!doc) return;
    setSaving(true);
    try {
      const body: Parameters<typeof api.documents.update>[2] = {};

      const newName = name.trim();
      if (newName !== (doc.name ?? "")) body.name = newName || undefined;

      const metadataUpdates: Record<string, unknown> = {};
      for (const field of metadataFields) {
        const raw = metaValues[field.name]?.trim();
        if (!raw) continue;
        metadataUpdates[field.name] = coerceForType(raw, field.type);
      }
      if (Object.keys(metadataUpdates).length > 0) body.metadata = metadataUpdates;

      if (Object.keys(body).length === 0) {
        onClose();
        return;
      }

      const updated = await api.documents.update(slug, doc.id, body);
      toast.success("Document updated");
      onSaved(updated);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Update failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={!!doc} onOpenChange={open => { if (!open) onClose(); }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit document</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSave} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label className="text-xs text-zinc-500">Display name</Label>
            <Input
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Optional display name"
            />
          </div>

          {metadataFields.length > 0 && (
            <div className="space-y-3">
              <p className="text-xs font-medium text-zinc-500">Properties</p>
              {metadataFields.map(field => (
                <div key={field.id} className="space-y-1.5">
                  <Label className="text-xs text-zinc-500 flex items-center gap-1.5">
                    {field.name}
                    <span className="text-zinc-300">{field.type}</span>
                  </Label>
                  <MetaFieldInput
                    field={field}
                    value={metaValues[field.name] ?? ""}
                    onChange={v => setField(field.name, v)}
                  />
                </div>
              ))}
            </div>
          )}

          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={saving}>{saving ? "Saving…" : "Save"}</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function coerceForType(raw: string, type: MetadataField["type"]): unknown {
  switch (type) {
    case "num":  return parseFloat(raw);
    case "bool": return raw === "true";
    case "arr":  return raw.split(",").map(s => s.trim()).filter(Boolean);
    default:     return raw;
  }
}
