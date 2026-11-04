import { useState, type FormEvent } from "react";
import { router } from "@inertiajs/react";
import { ArrowRight, Globe2, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { api } from "@/lib/api";
import type { Project } from "@/types";
export default function ProjectForm({
  project,
  small = false,
}: {
  project?: Project;
  small?: boolean;
}) {
  const [open, setOpen] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(e.currentTarget));
    setBusy(true);
    setError("");
    try {
      const p = await api<Project>("/api/projects", {
        ...data,
        id: String(project?.id || ""),
      });
      setOpen(false);
      router.visit(`/projects?project=${p.id}`);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant={project ? "outline" : "default"}
          size={small ? "sm" : "default"}
        >
          {project ? <Globe2 size={14} /> : <Plus size={15} />}{" "}
          {project ? "Editar projeto" : "Novo projeto"}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {project
              ? "Configurar projeto"
              : "Um novo projeto. Uma nova audiência."}
          </DialogTitle>
          <DialogDescription>
            Conecte um subdomínio e um remetente para sua newsletter.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} className="space-y-4 mt-2">
          <Field label="Nome do projeto">
            <Input
              name="name"
              required
              maxLength={100}
              placeholder="Hexascore"
              defaultValue={project?.name}
            />
          </Field>
          <Field label="Identificador">
            <Input
              name="slug"
              required
              maxLength={63}
              pattern="[a-z0-9]+(-[a-z0-9]+)*"
              placeholder="hexascore"
              defaultValue={project?.slug}
            />
          </Field>
          <Field label="Subdomínio">
            <Input
              name="domain"
              required
              placeholder="news.hexascore.app"
              defaultValue={project?.domain}
            />
          </Field>
          <Field label="E-mail remetente">
            <Input
              name="from_email"
              required
              type="email"
              placeholder="newsletter@hexascore.app"
              defaultValue={project?.from_email}
            />
            <p className="text-[11px] text-muted-foreground mt-1.5">
              O domínio do remetente precisa estar verificado no Resend.
            </p>
          </Field>
          <Field label="Descrição da newsletter">
            <Textarea
              name="description"
              maxLength={500}
              placeholder="Novidades, bastidores e tudo que estamos construindo."
              defaultValue={project?.description}
            />
          </Field>
          {error && (
            <p role="alert" className="text-xs text-red-400">
              {error}
            </p>
          )}
          <Button type="submit" disabled={busy} className="w-full">
            {busy
              ? "Salvando…"
              : project
                ? "Salvar alterações"
                : "Criar projeto"}
            <ArrowRight size={15} />
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
export function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <label className="block">
      <span className="block text-xs font-medium mb-2">{label}</span>
      {children}
    </label>
  );
}
