import { Head, Link } from "@inertiajs/react";
import { ArrowLeft, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
export default function ErrorPage({
  status,
  message,
}: {
  status: number;
  message: string;
}) {
  return (
    <>
      <Head title={`Error ${status}`} />
      <main className="auth-page">
        <section className="text-center max-w-md">
          <span className="brand-mark mx-auto mb-8">
            <Send size={17} />
          </span>
          <p className="eyebrow mb-3">ERROR {status}</p>
          <h1 className="text-2xl font-semibold tracking-tight">
            {status === 403
              ? "This access is unavailable."
              : status === 404
                ? "We could not find this page."
                : status >= 500
                  ? "Something went wrong."
                  : "We could not continue."}
          </h1>
          <p className="text-sm text-muted-foreground mt-4 leading-6">
            {message}
          </p>
          <Button asChild variant="outline" className="mt-7">
            <Link href="/">
              <ArrowLeft size={14} />
              Back to home
            </Link>
          </Button>
        </section>
      </main>
    </>
  );
}
