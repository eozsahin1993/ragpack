import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function friendlyUri(uri: string) {
  return uri.replace(/^upload:\/\//, "").replace(/^file:\/\//, "");
}

const MIME_LABELS: Record<string, string> = {
  "application/pdf":                                                                        "PDF",
  "text/html":                                                                              "HTML",
  "text/plain":                                                                             "Text",
  "text/markdown":                                                                          "Markdown",
  "text/csv":                                                                               "CSV",
  "application/json":                                                                       "JSON",
  "text/xml":                                                                               "XML",
  "application/xml":                                                                        "XML",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document":               "Word",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":                     "Excel",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation":             "PowerPoint",
};

export function friendlyMimeType(mime: string) {
  return MIME_LABELS[mime] ?? mime.split("/").pop() ?? mime;
}

export function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.floor(hrs / 24);
  if (days < 7) return `${days}d ago`;
  return new Date(dateStr).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}
