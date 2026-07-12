"use client";

import { useState } from "react";
import { Copy, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { CreatedApiKey } from "@/lib/api";

interface RevealKeyDialogProps {
  apiKey: CreatedApiKey | null;
  onClose: () => void;
}

export function RevealKeyDialog({ apiKey, onClose }: RevealKeyDialogProps) {
  const [copied, setCopied] = useState(false);

  async function copyKey() {
    if (!apiKey) return;
    await navigator.clipboard.writeText(apiKey.key);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <Dialog open={!!apiKey} onOpenChange={open => !open && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Key created</DialogTitle>
          <DialogDescription>
            Copy this key now — it won&apos;t be shown again. If you lose it, you&apos;ll need to create a new one.
          </DialogDescription>
        </DialogHeader>
        {apiKey && (
          <div className="space-y-4 pt-2">
            <div className="flex items-center gap-2">
              <code className="flex-1 text-sm bg-muted rounded-lg px-3 py-2 break-all">{apiKey.key}</code>
              <Button type="button" variant="outline" size="icon" onClick={copyKey}>
                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
              </Button>
            </div>
            <div className="flex justify-end">
              <Button type="button" onClick={onClose}>Done</Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
