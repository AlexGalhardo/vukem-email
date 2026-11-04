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
  pending: "Pending",
  confirmed: "Confirmed",
  unsubscribed: "Unsubscribed",
  draft: "Draft",
  queued: "Queued",
  completed: "Processed",
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
  return new Intl.DateTimeFormat("en-US", {
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
        "Could not copy. Select and copy the address manually.",
      );
    }
  }
  const title =
    path === "/subscribers"
      ? "Your audience."
      : path === "/campaigns"
        ? "Conversations that connect."
        : path === "/projects"
          ? "A place for every project."
          : path === "/settings"
            ? "Make it yours."
            : "Every connection matters.";
  const description =
    path === "/subscribers"
      ? "People who chose to follow what you are building."
      : path === "/campaigns"
        ? "One good idea, one message, and your next connection."
        : path === "/projects"
          ? "Manage your newsletters, domains, and senders in one place."
          : path === "/settings"
            ? "The connections behind your newsletter."
            : "A closer look at your newsletter and the people on the other side.";
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
            ? "Subscribers"
            : path === "/campaigns/new"
              ? "New campaign"
              : path === "/campaigns"
                ? "Campaigns"
                : path === "/projects"
                  ? "Projects"
                  : path === "/settings"
                    ? "Settings"
                    : "Overview"
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
                {path === "/" ? "OVERVIEW" : path.slice(1).toUpperCase()}
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
                        Public page
                        <ArrowUpRight size={14} />
                      </a>
                    </Button>
                    <Button asChild>
                      <Link href={link("/campaigns/new")}>
                        <Plus size={15} />
                        New campaign
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
                LET'S GET STARTED
              </Badge>
              <h2 className="text-2xl font-semibold tracking-tight">
                Your first newsletter starts here.
              </h2>
              <p className="text-sm text-muted-foreground max-w-md mx-auto mt-3 mb-7 leading-7">
                Create a project, connect your subdomain, and welcome people
                who want to follow your journey.
              </p>
              <ProjectForm />
              <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-8 max-w-2xl mx-auto text-left">
                {[
                  [
                    "01",
                    "Connect your project",
                    "Your own domain and sender.",
                  ],
                  [
                    "02",
                    "Invite your audience",
                    "A page ready to accept subscriptions.",
                  ],
                  ["03", "Start a conversation", "Create campaigns your way."],
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
                    label: "Confirmed subscribers",
                    value: stats.confirmed,
                    icon: Users,
                    caption: "Ready to receive your updates",
                    green: true,
                  },
                  {
                    label: "Awaiting confirmation",
                    value: stats.pending,
                    icon: Mail,
                    caption: "One confirmation to get started",
                  },
                  {
                    label: "Campaigns created",
                    value: stats.campaigns,
                    icon: Send,
                    caption: "Your stories reaching the world",
                  },
                  {
                    label: "Confirmation rate",
                    value: `${stats.confirmed + stats.pending ? Math.round((stats.confirmed / (stats.confirmed + stats.pending)) * 100) : 0}%`,
                    icon: ShieldCheck,
                    caption: "Among active and pending subscriptions",
                  },
                ].map((metric) => (
                  <div className="panel p-5" key={metric.label}>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>{metric.label}</span>
                      <metric.icon size={15} />
                    </div>
                    <div className="text-3xl font-semibold tracking-tight mt-5 tabular-nums">
                      {typeof metric.value === "number"
                        ? metric.value.toLocaleString("en-US")
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
                        Your audience, up close.
                      </h2>
                      <p className="text-xs text-muted-foreground mt-2">
                        Every subscription begins a connection.
                      </p>
                    </div>
                    <Badge variant="outline" className="text-[10px]">
                      All subscribers
                    </Badge>
                  </div>
                  <Audience stats={stats} />
                  <div className="grid grid-cols-3 border-t border-border pt-5 text-xs">
                    {[
                      ["confirmed", "Confirmed", "bg-emerald-400"],
                      ["pending", "Pending", "bg-amber-300"],
                      ["unsubscribed", "Unsubscribed", "bg-neutral-500"],
                    ].map(([key, label, color]) => (
                      <div key={key}>
                        <span className="flex gap-1.5 text-[10px] text-muted-foreground items-center">
                          <span className={`size-1.5 rounded-full ${color}`} />
                          {label}
                        </span>
                        <p className="text-lg font-medium mt-2">
                          {(stats[key] || 0).toLocaleString("en-US")}
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
                  <p className="eyebrow mb-3">YOUR NEXT EDITION</p>
                  <h2 className="text-xl tracking-tight font-medium max-w-64 leading-7">
                    Something new
                    <br />
                    to share?
                  </h2>
                  <p className="text-xs leading-6 text-muted-foreground mt-3 mb-7 max-w-64">
                    An update, an idea, or a look behind the scenes.
                    Your audience wants to hear it.
                  </p>
                  <Button
                    asChild
                    className="mt-auto self-start"
                    variant="outline"
                  >
                    <Link href={link("/campaigns/new")}>
                      Create a campaign
                      <ArrowRight size={14} />
                    </Link>
                  </Button>
                </section>
              </div>
              <div className="panel mb-6">
                <div className="panel-heading">
                  <div>
                    <h2 className="section-title">Recent subscriptions</h2>
                    <p className="text-[11px] text-muted-foreground mt-1.5">
                      The people who just joined.
                    </p>
                  </div>
                  <Link
                    href={link("/subscribers")}
                    className="text-xs text-muted-foreground flex gap-1.5 items-center hover:text-foreground"
                  >
                    View all
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
                    Your next subscriber may be one link away.
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
                  {copied ? "Copied" : "Copy link"}
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
                  <h2 className="section-title">All campaigns</h2>
                  <Badge variant="secondary">{campaigns.length}</Badge>
                </div>
                <span className="text-[11px] text-muted-foreground">
                  Latest 100 campaigns
                </span>
              </div>
              {campaigns.length ? (
                <div className="overflow-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="pl-6">Campaign</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Recipients</TableHead>
                        <TableHead>Accepted by provider</TableHead>
                        <TableHead>Failed / skipped</TableHead>
                        <TableHead>Date</TableHead>
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
                                  Edit
                                  <ArrowRight size={13} />
                                </Link>
                              </Button>
                            ) : (
                              <span className="text-[10px] text-muted-foreground">
                                {deliveryStats[c.id]?.queued || 0} queued
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
                  title="Your first campaign is ahead."
                  description="Share an update. Tell a story. Start a conversation with your subscribers."
                  icon={Send}
                />
              )}
              <div className="px-6 py-4 border-t border-border text-[10px] text-muted-foreground">
                Accepted by provider means Resend accepted the message; final
                delivery depends on the recipient's mail server.
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
                      "A space to connect your project with people."}
                  </p>
                  <div className="border-t border-border mt-5 pt-4 flex gap-2">
                    <Button asChild size="sm" variant="outline">
                      <Link href={`/?project=${p.id}`}>
                        Open project
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
                    Another project. New connections.
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
                <h2 className="section-title">Your project</h2>
                <dl className="mt-6 space-y-5 text-xs">
                  {[
                    ["Name", project.name],
                    ["Domain", project.domain],
                    ["Sender", project.from_email],
                    ["Public page", publicUrl],
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
                <h2 className="section-title">Access and delivery</h2>
                <div className="mt-6 space-y-6">
                  {[
                    {
                      icon: ShieldCheck,
                      title: "Google OAuth",
                      text: "Access restricted to the verified address aleexgvieira@gmail.com.",
                    },
                    {
                      icon: Send,
                      title: "Resend + React Email",
                      text: "Confirmations and campaigns use a persistent queue, automatic retries, and unsubscribe links.",
                    },
                    {
                      icon: Code2,
                      title: "Markdown, MDX, and components",
                      text: "GFM with tables, lists, code, and React Email components. MDX with variables, expressions, and components in an isolated interpreter.",
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
                    Invite people to your newsletter
                  </h2>
                </div>
                <p className="text-xs text-muted-foreground mt-3">
                  Add this link to your project's landing page. The form and
                  confirmation flow are ready on the public page.
                </p>
                <div className="mt-5 flex gap-3">
                  <code className="flex-1 text-xs bg-muted/40 border border-border rounded-lg p-4 break-all">{`<a href="${publicUrl}">Subscribe to newsletter</a>`}</code>
                  <Button
                    variant="outline"
                    onClick={() =>
                      copy(`<a href="${publicUrl}">Subscribe to newsletter</a>`)
                    }
                    aria-label="Copy HTML"
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
          {total.toLocaleString("en-US")}
        </span>
        <span className="text-xs text-muted-foreground">
          total subscriptions
        </span>
      </div>
      <div
        className="flex gap-1 h-8 rounded overflow-hidden mt-6 bg-muted/50"
        aria-label="Subscriber distribution"
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
          ? "Distribution by subscription status"
          : "Your first connections will appear here."}
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
            <TableHead className="pl-6">Email</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Subscribed on</TableHead>
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
                      Unsubscribe
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
      title="Your first connections are on their way."
      description="Share your subscription page to welcome new subscribers."
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
            aria-label="Search email"
            className="pl-9 h-9 w-60"
            placeholder="Search by email…"
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
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="confirmed">Confirmed</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="unsubscribed">Unsubscribed</SelectItem>
            </SelectContent>
          </Select>
          <Button asChild variant="outline" size="sm">
            <a href={`/api/subscribers/export?project=${project.id}`}>
              <ArrowDownToLine size={13} />
              Export CSV
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
            ? `${(page - 1) * 20 + 1}–${Math.min(page * 20, total)} of ${total}`
            : "0 subscribers"}
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
            Previous
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
            Next
            <ChevronRight size={13} />
          </Button>
        </div>
      </div>
    </div>
  );
}
