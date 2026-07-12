"use client";

import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { api, ApiKey, Collection, CreatedApiKey } from "@/lib/api";
import { PageHeader } from "@/components/page-header";
import { KeysTable } from "./_components/keys-table";
import { CreateKeyDialog } from "./_components/create-key-dialog";
import { RevealKeyDialog } from "./_components/reveal-key-dialog";

export default function KeysPage() {
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [revealKey, setRevealKey] = useState<CreatedApiKey | null>(null);

  async function load() {
    try {
      const [keysData, collectionsData] = await Promise.all([
        api.keys.list(),
        api.collections.list(),
      ]);
      setKeys(keysData.keys ?? []);
      setCollections(collectionsData.collections ?? []);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to load API keys");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function handleDelete(k: ApiKey) {
    if (!confirm(`Delete "${k.name}"? Anything using this key will immediately lose access.`)) return;
    try {
      await api.keys.delete(k.id);
      await load();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "Failed to delete API key");
    }
  }

  return (
    <div className="space-y-8">
      <PageHeader
        title="API Keys"
        description="Keys for authenticating external clients against the public API. Each key's access is scoped independently — per collection, and separately for instance administration."
        action={<Button className="gap-2" onClick={() => setCreateOpen(true)}><Plus className="w-4 h-4" /> New key</Button>}
      />

      <KeysTable keys={keys} collections={collections} loading={loading} onDelete={handleDelete} />

      <CreateKeyDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        collections={collections}
        onCreated={async created => {
          setRevealKey(created);
          await load();
        }}
      />

      <RevealKeyDialog apiKey={revealKey} onClose={() => setRevealKey(null)} />
    </div>
  );
}
