"use client";

import { useParams, useRouter } from "next/navigation";
import { DocumentView } from "@/components/documents/document-view";

export default function DocumentPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  return <DocumentView slug={null} id={id} onDeleted={() => router.push("/documents")} />;
}
