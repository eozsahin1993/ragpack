import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background text-foreground">
      <p className="text-6xl font-bold text-primary mb-4">404</p>
      <p className="text-lg text-muted-foreground mb-8">Page not found</p>
      <Link href="/" className="text-sm text-primary hover:underline">
        Go home
      </Link>
    </div>
  );
}
