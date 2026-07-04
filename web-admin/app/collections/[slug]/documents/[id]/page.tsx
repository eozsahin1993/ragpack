import { redirect } from "next/navigation";

export default function DocumentPage({ params }: { params: { slug: string; id: string } }) {
  redirect(`/collections/${params.slug}/documents/${params.id}/chunks`);
}
