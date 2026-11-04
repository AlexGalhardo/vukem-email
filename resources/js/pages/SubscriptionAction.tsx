import { Head } from "@inertiajs/react";
import { CheckCheck, MailCheck, Send } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
export default function SubscriptionAction({
  action,
  token,
  subscriber,
}: {
  action: "confirm" | "unsubscribe";
  token: string;
  subscriber: string;
}) {
  const [busy, setBusy] = useState(false),
    [message, setMessage] = useState(""),
    [error, setError] = useState("");
  const confirm = action === "confirm";
  async function submit() {
    setBusy(true);
    setError("");
    try {
      const r = await api(`/${action}`, { token, subscriber });
      setMessage(r.message);
      history.replaceState(null, "", location.pathname);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <Head title={confirm ? "Confirmar inscrição" : "Cancelar inscrição"} />
      <main className="auth-page">
        <section className="w-full max-w-md rounded-xl border border-border bg-card p-9 text-center">
          <div className="size-14 rounded-xl border border-border mx-auto mb-6 grid place-items-center">
            {message ? (
              <CheckCheck className="text-emerald-400" />
            ) : (
              <MailCheck />
            )}
          </div>
          <h1 className="text-2xl font-semibold tracking-tight">
            {message
              ? "Tudo certo."
              : confirm
                ? "Confirme sua inscrição."
                : "Vamos sentir sua falta."}
          </h1>
          <p className="text-sm text-muted-foreground mt-4 leading-6">
            {message ||
              (confirm
                ? "Clique abaixo para confirmar seu e-mail e começar a receber as novidades."
                : "Confirme abaixo para parar de receber esta newsletter.")}
          </p>
          {!message && (
            <Button
              className="mt-7 w-full"
              disabled={busy || !token}
              onClick={submit}
            >
              {busy
                ? "Aguarde…"
                : confirm
                  ? "Confirmar meu e-mail"
                  : "Cancelar minha inscrição"}
            </Button>
          )}
          {error && (
            <p role="alert" className="text-xs text-red-400 mt-4">
              {error}
            </p>
          )}
        </section>
      </main>
    </>
  );
}
