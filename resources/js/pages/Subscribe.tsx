import { Head } from "@inertiajs/react";
import { ArrowRight, Check, CheckCheck, Send } from "lucide-react";
import { useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import type { Project } from "@/types";
export default function Subscribe({ project }: { project: Project }) {
  const [email, setEmail] = useState(""),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [message, setMessage] = useState("");
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const form = new FormData(e.currentTarget);
      const result = await api(`/subscribe/${project.slug}`, {
        email,
        website: form.get("website"),
      });
      setMessage(result.message);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <Head title={`${project.name} · Newsletter`} />
      <main className="auth-page">
        <header className="absolute inset-x-0 top-0 px-7 md:px-14 h-24 flex items-center justify-between">
          <div className="flex items-center gap-2.5 font-semibold">
            <span className="brand-mark">
              <Send size={16} />
            </span>
            {project.name}
          </div>
          <span className="text-xs text-muted-foreground">
            One letter. Many possibilities.
          </span>
        </header>
        <section className="max-w-xl w-full text-center relative">
          <span className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground mb-7">
            <span className="size-1.5 bg-emerald-400 rounded-full" />A
            {project.name} newsletter
          </span>
          <h1 className="text-5xl md:text-6xl font-semibold tracking-[-.05em] leading-[1.1]">
            The next chapter,
            <br />
            <span className="text-muted-foreground">
              in your inbox.
            </span>
          </h1>
          <p className="text-muted-foreground mt-6 max-w-md mx-auto text-sm md:text-base leading-7">
            {project.description ||
              `Updates, behind-the-scenes stories, and ideas from ${project.name}. Subscribe to follow what we are building.`}
          </p>
          {message ? (
            <div
              role="status"
              className="rounded-xl border border-emerald-400/20 bg-emerald-400/5 p-6 mt-8 text-sm leading-6"
            >
              <CheckCheck className="mx-auto mb-3 text-emerald-400" />
              {message}
            </div>
          ) : (
            <form onSubmit={submit} className="mt-9 max-w-md mx-auto">
              <label htmlFor="email" className="sr-only">
                Your email
              </label>
              <div className="flex gap-2">
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Your email address"
                  className="h-12 bg-card"
                />
                <Button disabled={busy} type="submit" className="h-12 px-5">
                  {busy ? "Subscribing…" : "Subscribe"}
                  <ArrowRight size={15} />
                </Button>
              </div>
              <input
                tabIndex={-1}
                aria-hidden="true"
                autoComplete="off"
                name="website"
                className="hidden"
              />
              {error && (
                <p role="alert" className="text-red-400 text-xs mt-3">
                  {error}
                </p>
              )}
              <div className="flex justify-center gap-5 text-[11px] text-muted-foreground mt-5">
                <span className="flex items-center gap-1">
                  <Check size={12} />
                  No spam. Just the essentials.
                </span>
                <span className="flex items-center gap-1">
                  <Check size={12} />
                  Unsubscribe anytime.
                </span>
              </div>
              <p className="text-[11px] text-muted-foreground/70 mt-5">
                By subscribing, you agree to receive the newsletter from{" "}
                {project.name}.<br />
                We will email you to confirm your subscription.
              </p>
            </form>
          )}
        </section>
        <footer className="absolute bottom-7 text-[11px] text-muted-foreground">
          {project.name} <span className="mx-2">/</span> Powered by Vukem Email
        </footer>
      </main>
    </>
  );
}
