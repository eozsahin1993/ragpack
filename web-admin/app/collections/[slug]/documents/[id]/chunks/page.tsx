"use client";

import { useParams, useRouter } from "next/navigation";
import { DocumentView } from "@/components/documents/document-view";

export default function ChunksPage() {
  const { slug, id } = useParams<{ slug: string; id: string }>();
  const router = useRouter();

  return <DocumentView slug={slug} id={id} onDeleted={() => router.push(`/collections/${slug}`)} />;
}
