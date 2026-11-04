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
      <Head title={confirm ? "Confirm subscription" : "Unsubscribe"} />
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
              ? "All set."
              : confirm
                ? "Confirm your subscription."
                : "We will miss you."}
          </h1>
          <p className="text-sm text-muted-foreground mt-4 leading-6">
            {message ||
              (confirm
                ? "Click below to confirm your email and start receiving updates."
                : "Confirm below to stop receiving this newsletter.")}
          </p>
          {!message && (
            <Button
              className="mt-7 w-full"
              disabled={busy || !token}
              onClick={submit}
            >
              {busy
                ? "Please wait…"
                : confirm
                  ? "Confirm my email"
                  : "Unsubscribe"}
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
