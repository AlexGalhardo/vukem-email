import { Head, Link, router, usePage } from "@inertiajs/react";
import {
  ArrowDownToLine,
  ArrowRight,
  ArrowUpRight,
  Check,
  ChevronLeft,
  ChevronRight,
  Clipboard,
  Code2,
  ExternalLink,
  Globe2,
  Inbox,
  Mail,
  Plus,
  Search,
  Send,
  ShieldCheck,
  Sparkles,
  Users,
} from "lucide-react";
import { useEffect, useState, type FormEvent } from "react";
import AppLayout from "@/layouts/AppLayout";
import Composer from "@/components/Composer";
import ProjectForm from "@/components/ProjectForm";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api } from "@/lib/api";
import type { Campaign, DashboardProps, Project, Subscriber } from "@/types";
const labels: Record<string, string> = {
  pending: "Pendente",
  confirmed: "Confirmado",
  unsubscribed: "Descadastrado",
  draft: "Rascunho",
  queued: "Na fila",
  completed: "Processada",
};
function Status({ value }: { value: string }) {
  return (
    <Badge variant="outline" className={`status status-${value}`}>
      <span className="size-1 rounded-full bg-current" />
      {labels[value] || value}
    </Badge>
  );
}
function date(value: string) {
  return new Intl.DateTimeFormat("pt-BR", {
    day: "2-digit",
    month: "short",
  }).format(new Date(value));
}
function Empty({
  title,
  description,
  icon: Icon = Inbox,
}: {
  title: string;
  description: string;
  icon?: typeof Inbox;
}) {
  return (
    <div className="py-16 px-5 text-center">
      <div className="rounded-xl border border-border bg-muted/40 size-12 mx-auto grid place-items-center mb-4">
        <Icon size={20} className="text-muted-foreground" />
      </div>
      <p className="text-sm font-medium">{title}</p>
      <p className="text-xs text-muted-foreground mt-2 leading-6 max-w-sm mx-auto">
        {description}
      </p>
    </div>
  );
}
export default function Dashboard(props: DashboardProps) {
  const {
    project,
    projects,
    stats,
    campaigns,
    subscribers,
    total,
    page,
    filters,
    deliveryStats,
  } = props;
  const url = usePage().url,
    path = url.split("?")[0],
    query = new URLSearchParams(url.split("?")[1]);
  const [copied, setCopied] = useState(false),
    [error, setError] = useState("");
  const hasProject = !!project.id;
  const link = (href: string) => `${href}?project=${project.id}`;
  const publicUrl = `https://${project.domain}/subscribe`;
  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setError(
        "Não foi possível copiar. Selecione e copie o endereço manualmente.",
      );
    }
  }
  const title =
    path === "/subscribers"
      ? "Sua audiência."
      : path === "/campaigns"
        ? "Conversas que conectam."
        : path === "/projects"
          ? "Um lugar para cada projeto."
          : path === "/settings"
            ? "Do seu jeito."
            : "Cada conexão importa.";
  const description =
    path === "/subscribers"
      ? "Pessoas que escolheram acompanhar o que você está construindo."
      : path === "/campaigns"
        ? "Uma boa ideia, uma mensagem e sua próxima conexão."
        : path === "/projects"
          ? "Gerencie suas newsletters, domínios e remetentes em um só lugar."
          : path === "/settings"
            ? "As conexões que fazem sua newsletter acontecer."
            : "Uma visão de perto da sua newsletter e das pessoas do outro lado.";
  useEffect(() => {
    if (path === "/campaigns" && campaigns.some((c) => c.status === "queued")) {
      const id = setInterval(
        () => router.reload({ only: ["campaigns", "deliveryStats"] }),
        15000,
      );
      return () => clearInterval(id);
    }
  }, [path, campaigns]);
  return (
    <AppLayout project={project} projects={projects}>
      <Head
        title={
          path === "/subscribers"
            ? "Inscritos"
            : path === "/campaigns/new"
              ? "Nova campanha"
              : path === "/campaigns"
                ? "Campanhas"
                : path === "/projects"
                  ? "Projetos"
                  : path === "/settings"
                    ? "Configurações"
                    : "Visão geral"
        }
      />
      {path === "/campaigns/new" && hasProject ? (
        <Composer
          key={`${project.id}-${query.get("campaign") || "new"}`}
          project={project}
          campaign={campaigns.find(
            (c) =>
              c.id === Number(query.get("campaign")) && c.status === "draft",
          )}
          confirmed={stats.confirmed || 0}
        />
      ) : (
        <>
          <div className="flex gap-5 flex-wrap justify-between items-start mb-8">
            <div>
              <p className="eyebrow mb-2.5">
                {path === "/" ? "VISÃO GERAL" : path.slice(1).toUpperCase()}
              </p>
              <h1 className="page-title">{title}</h1>
              <p className="page-description">{description}</p>
            </div>
            <div className="flex gap-2 mt-3">
              {path === "/projects" ? (
                <ProjectForm />
              ) : (
                hasProject &&
                path !== "/settings" && (
                  <>
                    <Button
                      asChild
                      variant="outline"
                      className="hidden sm:inline-flex"
                    >
                      <a href={publicUrl} target="_blank" rel="noreferrer">
                        Página pública
                        <ArrowUpRight size={14} />
                      </a>
                    </Button>
                    <Button asChild>
                      <Link href={link("/campaigns/new")}>
                        <Plus size={15} />
                        Nova campanha
                      </Link>
                    </Button>
                  </>
                )
              )}
            </div>
          </div>
          {error && (
            <div role="alert" className="notice error mb-5">
              {error}
            </div>
          )}
          {!hasProject && path !== "/projects" && (
            <div className="panel p-10 md:p-16 text-center">
              <div className="size-16 rounded-2xl bg-muted mx-auto grid place-items-center mb-6">
                <Globe2 size={27} />
              </div>
              <Badge variant="outline" className="mb-5">
                VAMOS COMEÇAR
              </Badge>
              <h2 className="text-2xl font-semibold tracking-tight">
                Sua primeira newsletter começa aqui.
              </h2>
              <p className="text-sm text-muted-foreground max-w-md mx-auto mt-3 mb-7 leading-7">
                Crie um projeto, conecte seu subdomínio e abra espaço para as
                pessoas que querem acompanhar sua jornada.
              </p>
              <ProjectForm />
              <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-8 max-w-2xl mx-auto text-left">
                {[
                  [
                    "01",
                    "Conecte seu projeto",
                    "Um domínio e um remetente próprios.",
                  ],
                  [
                    "02",
                    "Convide sua audiência",
                    "Uma página pronta para receber inscrições.",
                  ],
                  ["03", "Comece uma conversa", "Crie campanhas do seu jeito."],
                ].map(([num, title, desc]) => (
                  <div key={num}>
                    <p className="font-mono text-[10px] text-muted-foreground mb-3">
                      {num} ──
                    </p>
                    <p className="text-xs font-medium">{title}</p>
                    <p className="text-[11px] text-muted-foreground mt-2 leading-5">
                      {desc}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          )}
          {hasProject && path === "/" && (
            <>
              <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 mb-6">
                {[
                  {
                    label: "Inscritos confirmados",
                    value: stats.confirmed,
                    icon: Users,
                    caption: "Prontos para receber suas novidades",
                    green: true,
                  },
                  {
                    label: "Aguardando confirmação",
                    value: stats.pending,
                    icon: Mail,
                    caption: "Uma confirmação para começar",
                  },
                  {
                    label: "Campanhas criadas",
                    value: stats.campaigns,
                    icon: Send,
                    caption: "Suas histórias ganhando o mundo",
                  },
                  {
                    label: "Taxa de confirmação",
                    value: `${stats.confirmed + stats.pending ? Math.round((stats.confirmed / (stats.confirmed + stats.pending)) * 100) : 0}%`,
                    icon: ShieldCheck,
                    caption: "Entre inscrições ativas e pendentes",
                  },
                ].map((metric) => (
                  <div className="panel p-5" key={metric.label}>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>{metric.label}</span>
                      <metric.icon size={15} />
                    </div>
                    <div className="text-3xl font-semibold tracking-tight mt-5 tabular-nums">
                      {typeof metric.value === "number"
                        ? metric.value.toLocaleString("pt-BR")
                        : metric.value}
                    </div>
                    <p className="text-[10px] text-muted-foreground mt-3 flex items-center gap-1.5">
                      {metric.green && (
                        <span className="size-1.5 rounded-full bg-emerald-400" />
                      )}
                      {metric.caption}
                    </p>
                  </div>
                ))}
              </div>
              <div className="grid grid-cols-1 xl:grid-cols-[1.6fr_1fr] gap-5 mb-6">
                <section className="panel p-6">
                  <div className="flex justify-between items-start">
                    <div>
                      <h2 className="section-title">
                        Sua audiência, de perto.
                      </h2>
                      <p className="text-xs text-muted-foreground mt-2">
                        Cada inscrição é o começo de uma conexão.
                      </p>
                    </div>
                    <Badge variant="outline" className="text-[10px]">
                      Todos os inscritos
                    </Badge>
                  </div>
                  <Audience stats={stats} />
                  <div className="grid grid-cols-3 border-t border-border pt-5 text-xs">
                    {[
                      ["confirmed", "Confirmados", "bg-emerald-400"],
                      ["pending", "Pendentes", "bg-amber-300"],
                      ["unsubscribed", "Descadastros", "bg-neutral-500"],
                    ].map(([key, label, color]) => (
                      <div key={key}>
                        <span className="flex gap-1.5 text-[10px] text-muted-foreground items-center">
                          <span className={`size-1.5 rounded-full ${color}`} />
                          {label}
                        </span>
                        <p className="text-lg font-medium mt-2">
                          {(stats[key] || 0).toLocaleString("pt-BR")}
                        </p>
                      </div>
                    ))}
                  </div>
                </section>
                <section className="panel p-6 flex flex-col relative overflow-hidden">
                  <span className="absolute -right-8 -top-8 rounded-full size-36 border border-border/70" />
                  <span className="absolute -right-2 -top-2 rounded-full size-24 border border-border/70" />
                  <div className="size-9 rounded-lg border border-border grid place-items-center mb-6">
                    <Send size={17} />
                  </div>
                  <p className="eyebrow mb-3">SUA PRÓXIMA EDIÇÃO</p>
                  <h2 className="text-xl tracking-tight font-medium max-w-64 leading-7">
                    Tem algo novo
                    <br />
                    para compartilhar?
                  </h2>
                  <p className="text-xs leading-6 text-muted-foreground mt-3 mb-7 max-w-64">
                    Uma atualização, uma ideia ou os bastidores do seu projeto.
                    Sua audiência quer saber.
                  </p>
                  <Button
                    asChild
                    className="mt-auto self-start"
                    variant="outline"
                  >
                    <Link href={link("/campaigns/new")}>
                      Criar uma campanha
                      <ArrowRight size={14} />
                    </Link>
                  </Button>
                </section>
              </div>
              <div className="panel mb-6">
                <div className="panel-heading">
                  <div>
                    <h2 className="section-title">Últimas inscrições</h2>
                    <p className="text-[11px] text-muted-foreground mt-1.5">
                      As pessoas que acabaram de chegar.
                    </p>
                  </div>
                  <Link
                    href={link("/subscribers")}
                    className="text-xs text-muted-foreground flex gap-1.5 items-center hover:text-foreground"
                  >
                    Ver todos
                    <ArrowRight size={13} />
                  </Link>
                </div>
                <SubscriberTable subscribers={subscribers.slice(0, 5)} />
              </div>
              <div className="panel p-5 flex flex-wrap gap-4 items-center">
                <div className="size-10 rounded-lg border border-border grid place-items-center">
                  <Globe2 size={19} />
                </div>
                <div className="flex-1">
                  <p className="text-xs font-medium">
                    Seu próximo inscrito pode estar a um link de distância.
                  </p>
                  <a
                    href={publicUrl}
                    className="text-[11px] text-muted-foreground mt-1.5 inline-block"
                    target="_blank"
                    rel="noreferrer"
                  >
                    {publicUrl}
                  </a>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => copy(publicUrl)}
                >
                  {copied ? <Check size={13} /> : <Clipboard size={13} />}{" "}
                  {copied ? "Copiado" : "Copiar link"}
                </Button>
              </div>
            </>
          )}
          {hasProject && path === "/subscribers" && <Subscribers {...props} />}{" "}
          {hasProject && path === "/campaigns" && (
            <div className="panel">
              <div className="panel-heading">
                <div className="flex items-center gap-2">
                  <Mail size={15} />
                  <h2 className="section-title">Todas as campanhas</h2>
                  <Badge variant="secondary">{campaigns.length}</Badge>
                </div>
                <span className="text-[11px] text-muted-foreground">
                  Últimas 100 campanhas
                </span>
              </div>
              {campaigns.length ? (
                <div className="overflow-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="pl-6">Campanha</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Destinatários</TableHead>
                        <TableHead>Aceitos pelo provedor</TableHead>
                        <TableHead>Falhas / ignorados</TableHead>
                        <TableHead>Data</TableHead>
                        <TableHead />
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {campaigns.map((c) => (
                        <TableRow key={c.id}>
                          <TableCell className="pl-6 py-5">
                            <div className="font-medium text-xs">
                              {c.subject}
                            </div>
                            <p className="text-[10px] text-muted-foreground mt-1.5">
                              {c.mode === "components"
                                ? "React Email"
                                : c.mode.toUpperCase()}
                            </p>
                          </TableCell>
                          <TableCell>
                            <Status value={c.status} />
                          </TableCell>
                          <TableCell>{c.recipients || "—"}</TableCell>
                          <TableCell>
                            {deliveryStats[c.id]?.accepted || "—"}
                          </TableCell>
                          <TableCell className="text-muted-foreground">
                            {deliveryStats[c.id]?.failed || 0} /{" "}
                            {deliveryStats[c.id]?.skipped || 0}
                          </TableCell>
                          <TableCell className="text-muted-foreground text-xs whitespace-nowrap">
                            {date(c.created_at)}
                          </TableCell>
                          <TableCell>
                            {c.status === "draft" ? (
                              <Button asChild size="sm" variant="ghost">
                                <Link
                                  href={`${link("/campaigns/new")}&campaign=${c.id}`}
                                >
                                  Editar
                                  <ArrowRight size={13} />
                                </Link>
                              </Button>
                            ) : (
                              <span className="text-[10px] text-muted-foreground">
                                {deliveryStats[c.id]?.queued || 0} na fila
                              </span>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <Empty
                  title="Sua primeira campanha está por vir."
                  description="Compartilhe uma novidade. Conte uma história. Comece uma conversa com seus inscritos."
                  icon={Send}
                />
              )}
              <div className="px-6 py-4 border-t border-border text-[10px] text-muted-foreground">
                Aceito pelo provedor indica que o Resend aceitou a mensagem; a
                entrega final depende do servidor do destinatário.
              </div>
            </div>
          )}
          {path === "/projects" && (
            <div className="grid md:grid-cols-2 xl:grid-cols-3 gap-5">
              {projects.map((p) => (
                <section className="panel p-5" key={p.id}>
                  <div className="flex justify-between items-center">
                    <span className="size-10 rounded-lg bg-muted border border-border grid place-items-center font-medium">
                      {p.name[0]}
                    </span>
                    <Globe2 size={17} className="text-muted-foreground" />
                  </div>
                  <h2 className="text-lg font-medium mt-5 tracking-tight">
                    {p.name}
                  </h2>
                  <p className="text-[11px] text-muted-foreground mt-1.5">
                    {p.domain}
                  </p>
                  <p className="text-xs text-muted-foreground mt-4 min-h-12 leading-6">
                    {p.description ||
                      "Um espaço para conectar seu projeto às pessoas."}
                  </p>
                  <div className="border-t border-border mt-5 pt-4 flex gap-2">
                    <Button asChild size="sm" variant="outline">
                      <Link href={`/?project=${p.id}`}>
                        Abrir projeto
                        <ArrowUpRight size={12} />
                      </Link>
                    </Button>
                    <ProjectForm project={p} small />
                  </div>
                </section>
              ))}
              <div className="rounded-xl border border-dashed border-border min-h-64 grid place-items-center">
                <div className="text-center">
                  <span className="size-10 border border-border rounded-full grid place-items-center mx-auto text-muted-foreground">
                    <Plus size={18} />
                  </span>
                  <p className="text-xs text-muted-foreground mt-4">
                    Mais um projeto. Novas conexões.
                  </p>
                  <div
                    id="add-project-trigger"
                    className="mt-4"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <ProjectForm small />
                  </div>
                </div>
              </div>
            </div>
          )}
          {hasProject && path === "/settings" && (
            <div className="grid lg:grid-cols-2 gap-5">
              <section className="panel p-6">
                <h2 className="section-title">Seu projeto</h2>
                <dl className="mt-6 space-y-5 text-xs">
                  {[
                    ["Nome", project.name],
                    ["Domínio", project.domain],
                    ["Remetente", project.from_email],
                    ["Página pública", publicUrl],
                  ].map(([k, v]) => (
                    <div key={k} className="flex justify-between gap-3">
                      <dt className="text-muted-foreground">{k}</dt>
                      <dd className="text-right break-all">{v}</dd>
                    </div>
                  ))}
                </dl>
                <div className="mt-6 border-t border-border pt-5">
                  <ProjectForm project={project} />
                </div>
              </section>
              <section className="panel p-6">
                <h2 className="section-title">Acesso e envio</h2>
                <div className="mt-6 space-y-6">
                  {[
                    {
                      icon: ShieldCheck,
                      title: "Google OAuth",
                      text: "Acesso exclusivo a aleexgvieira@gmail.com, com e-mail verificado.",
                    },
                    {
                      icon: Send,
                      title: "Resend + React Email",
                      text: "Confirmações e campanhas com fila persistente, tentativas automáticas e descadastro.",
                    },
                    {
                      icon: Code2,
                      title: "Markdown, MDX e componentes",
                      text: "GFM com tabelas, listas, código e componentes React Email. MDX com variáveis, expressões e componentes em um interpretador isolado.",
                    },
                  ].map((item) => (
                    <div key={item.title} className="flex gap-3">
                      <item.icon
                        size={17}
                        className="mt-0.5 text-muted-foreground"
                      />
                      <div>
                        <p className="text-xs font-medium">{item.title}</p>
                        <p className="text-xs text-muted-foreground mt-1.5 leading-6">
                          {item.text}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </section>
              <section className="panel p-6 lg:col-span-2">
                <div className="flex gap-3 items-center">
                  <Code2 size={17} />
                  <h2 className="section-title">
                    Convide pessoas para sua newsletter
                  </h2>
                </div>
                <p className="text-xs text-muted-foreground mt-3">
                  Adicione este link à landing page do seu projeto. O formulário
                  e a confirmação já estão prontos na página pública.
                </p>
                <div className="mt-5 flex gap-3">
                  <code className="flex-1 text-xs bg-muted/40 border border-border rounded-lg p-4 break-all">{`<a href="${publicUrl}">Assinar newsletter</a>`}</code>
                  <Button
                    variant="outline"
                    onClick={() =>
                      copy(`<a href="${publicUrl}">Assinar newsletter</a>`)
                    }
                    aria-label="Copiar HTML"
                  >
                    {copied ? <Check size={15} /> : <Clipboard size={15} />}
                  </Button>
                </div>
              </section>
            </div>
          )}
        </>
      )}
    </AppLayout>
  );
}
function Audience({ stats }: { stats: Record<string, number> }) {
  const total =
    (stats.confirmed || 0) + (stats.pending || 0) + (stats.unsubscribed || 0);
  return (
    <div className="py-10">
      <div className="flex items-baseline gap-2">
        <span className="text-4xl font-semibold tracking-tight">
          {total.toLocaleString("pt-BR")}
        </span>
        <span className="text-xs text-muted-foreground">
          inscrições no total
        </span>
      </div>
      <div
        className="flex gap-1 h-8 rounded overflow-hidden mt-6 bg-muted/50"
        aria-label="Distribuição dos inscritos"
      >
        {total ? (
          ["confirmed", "pending", "unsubscribed"].map((key, i) => (
            <div
              key={key}
              title={`${labels[key]}: ${stats[key]}`}
              style={{ width: `${(stats[key] / total) * 100}%` }}
              className={
                ["bg-emerald-400/85", "bg-amber-300/80", "bg-neutral-500"][i]
              }
            />
          ))
        ) : (
          <div className="w-full h-full audience-empty" />
        )}
      </div>
      <div className="text-[10px] text-muted-foreground mt-3">
        {total
          ? "Distribuição por status de inscrição"
          : "As primeiras conexões vão aparecer aqui."}
      </div>
    </div>
  );
}
function SubscriberTable({
  subscribers,
  actions = false,
  onAction,
}: {
  subscribers: Subscriber[];
  actions?: boolean;
  onAction?: (s: Subscriber) => void;
}) {
  return subscribers.length ? (
    <div className="overflow-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="pl-6">E-mail</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Inscrito em</TableHead>
            {actions && <TableHead />}
          </TableRow>
        </TableHeader>
        <TableBody>
          {subscribers.map((s) => (
            <TableRow key={s.id}>
              <TableCell className="pl-6 py-4">
                <div className="flex items-center gap-3">
                  <div className="size-7 rounded-full bg-muted border border-border text-[10px] uppercase grid place-items-center text-muted-foreground">
                    {s.email.slice(0, 2)}
                  </div>
                  <span className="text-xs">{s.email}</span>
                </div>
              </TableCell>
              <TableCell>
                <Status value={s.status} />
              </TableCell>
              <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                {date(s.created_at)}
              </TableCell>
              {actions && (
                <TableCell className="text-right pr-6">
                  {s.status !== "unsubscribed" && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-xs text-muted-foreground"
                      onClick={() => onAction?.(s)}
                    >
                      Descadastrar
                    </Button>
                  )}
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  ) : (
    <Empty
      title="As primeiras conexões vêm aí."
      description="Compartilhe sua página de inscrição para começar a receber novos inscritos."
      icon={Users}
    />
  );
}
function Subscribers(props: DashboardProps) {
  const { project, subscribers, total, page, filters } = props;
  const [search, setSearch] = useState(filters.q),
    [error, setError] = useState("");
  useEffect(() => setSearch(filters.q), [filters.q]);
  function filter(status: string) {
    router.get(
      "/subscribers",
      {
        project: project.id,
        q: search,
        status: status === "all" ? "" : status,
      },
      { preserveState: true },
    );
  }
  async function unsubscribe(s: Subscriber) {
    try {
      await api(`/api/subscribers/${s.id}/unsubscribe`, {});
      router.reload();
    } catch (e) {
      setError((e as Error).message);
    }
  }
  return (
    <div className="panel">
      <div className="panel-heading flex-wrap">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            filter(filters.status || "all");
          }}
          className="relative"
        >
          <Search
            className="absolute top-2.5 left-3 text-muted-foreground"
            size={14}
          />
          <Input
            aria-label="Buscar e-mail"
            className="pl-9 h-9 w-60"
            placeholder="Buscar por e-mail…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </form>
        <div className="flex items-center gap-2">
          <Select value={filters.status || "all"} onValueChange={filter}>
            <SelectTrigger className="w-40 h-9 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Todos os status</SelectItem>
              <SelectItem value="confirmed">Confirmados</SelectItem>
              <SelectItem value="pending">Pendentes</SelectItem>
              <SelectItem value="unsubscribed">Descadastrados</SelectItem>
            </SelectContent>
          </Select>
          <Button asChild variant="outline" size="sm">
            <a href={`/api/subscribers/export?project=${project.id}`}>
              <ArrowDownToLine size={13} />
              Exportar CSV
            </a>
          </Button>
        </div>
      </div>
      {error && (
        <div className="notice error m-4" role="alert">
          {error}
        </div>
      )}
      <SubscriberTable
        subscribers={subscribers}
        actions
        onAction={unsubscribe}
      />
      <div className="px-6 py-4 border-t border-border flex items-center justify-between">
        <span className="text-[11px] text-muted-foreground">
          {total
            ? `${(page - 1) * 20 + 1}–${Math.min(page * 20, total)} de ${total}`
            : "0 inscritos"}
        </span>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={page <= 1}
            onClick={() =>
              router.get("/subscribers", {
                project: project.id,
                ...filters,
                page: page - 1,
              })
            }
          >
            <ChevronLeft size={13} />
            Anterior
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={page * 20 >= total}
            onClick={() =>
              router.get("/subscribers", {
                project: project.id,
                ...filters,
                page: page + 1,
              })
            }
          >
            Próxima
            <ChevronRight size={13} />
          </Button>
        </div>
      </div>
    </div>
  );
}
