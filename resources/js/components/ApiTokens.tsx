import { router } from "@inertiajs/react";
import { KeyRound } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import type { ApiToken, Project } from "@/types";

const abilities: [string, string][] = [
  ["subscribers:write", "Create subscribers and unsubscribe them"],
  ["campaigns:send", "Send draft campaigns"],
];

export default function ApiTokens({
  project,
  tokens,
}: {
  project: Project;
  tokens: ApiToken[];
}) {
  const [created, setCreated] = useState(""),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    setBusy(true);
    setError("");
    try {
      const result = await api<{ token: string }>(
        `/api/projects/${project.id}/tokens`,
        {
          name: String(form.get("name") || ""),
          abilities: form.getAll("abilities").join(","),
        },
      );
      setCreated(result.token);
      router.reload({ only: ["apiTokens"] });
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function revoke(id: string) {
    setError("");
    try {
      await api(`/api/projects/${project.id}/tokens/${id}/revoke`, {});
      router.reload({ only: ["apiTokens"] });
    } catch (e) {
      setError((e as Error).message);
    }
  }
  return (
    <section className="panel p-6 lg:col-span-2">
      <div className="flex gap-3 items-center">
        <KeyRound size={17} />
        <h2 className="section-title">API tokens</h2>
      </div>
      <p className="text-xs text-muted-foreground mt-3">
        Tokens let another server call <code>/api/v1</code> for this project
        only. The value is shown once; only its hash is stored.
      </p>
      {created && (
        <code className="block mt-4 text-xs bg-muted/40 border border-border rounded-lg p-4 break-all">
          {created}
        </code>
      )}
      <ul className="mt-5 space-y-3 text-xs">
        {tokens.map((t) => (
          <li key={t.id} className="flex items-center justify-between gap-3">
            <span>
              <span className="font-medium">{t.name}</span>{" "}
              <span className="text-muted-foreground">
                {t.abilities.join(", ")}
                {t.last_used_at ? "" : " (never used)"}
              </span>
            </span>
            <Button
              variant="outline"
              size="sm"
              type="button"
              onClick={() => revoke(t.id)}
            >
              Revoke
            </Button>
          </li>
        ))}
        {!tokens.length && (
          <li className="text-muted-foreground">No active tokens.</li>
        )}
      </ul>
      <form
        onSubmit={create}
        className="mt-5 flex flex-wrap items-center gap-4"
      >
        <Input
          name="name"
          placeholder="Token name"
          aria-label="Token name"
          required
          maxLength={100}
          className="max-w-56"
        />
        {abilities.map(([value, text]) => (
          <span key={value} className="flex items-center gap-2 text-xs">
            <input
              type="checkbox"
              name="abilities"
              value={value}
              id={`ability-${value}`}
              defaultChecked
            />
            <label htmlFor={`ability-${value}`}>{text}</label>
          </span>
        ))}
        <Button type="submit" disabled={busy}>
          Create token
        </Button>
      </form>
      {error && <p className="mt-3 text-xs text-destructive">{error}</p>}
    </section>
  );
}
