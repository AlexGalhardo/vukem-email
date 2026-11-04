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
            Uma carta. Muitas possibilidades.
          </span>
        </header>
        <section className="max-w-xl w-full text-center relative">
          <span className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground mb-7">
            <span className="size-1.5 bg-emerald-400 rounded-full" />A
            newsletter de {project.name}
          </span>
          <h1 className="text-5xl md:text-6xl font-semibold tracking-[-.05em] leading-[1.1]">
            O próximo capítulo,
            <br />
            <span className="text-muted-foreground">
              na sua caixa de entrada.
            </span>
          </h1>
          <p className="text-muted-foreground mt-6 max-w-md mx-auto text-sm md:text-base leading-7">
            {project.description ||
              `Novidades, bastidores e ideias de ${project.name}. Assine para acompanhar o que estamos construindo.`}
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
                Seu e-mail
              </label>
              <div className="flex gap-2">
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Seu melhor e-mail"
                  className="h-12 bg-card"
                />
                <Button disabled={busy} type="submit" className="h-12 px-5">
                  {busy ? "Inscrevendo…" : "Quero receber"}
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
                  Sem spam. Só o que importa.
                </span>
                <span className="flex items-center gap-1">
                  <Check size={12} />
                  Cancele quando quiser.
                </span>
              </div>
              <p className="text-[11px] text-muted-foreground/70 mt-5">
                Ao assinar, você concorda em receber a newsletter de{" "}
                {project.name}.<br />
                Enviaremos um e-mail para confirmar sua inscrição.
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
