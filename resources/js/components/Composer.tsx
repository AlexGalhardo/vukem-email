import { useRef, useState } from "react";
import { router } from "@inertiajs/react";
import {
  ArrowLeft,
  Check,
  Code2,
  Eye,
  FileText,
  LayoutTemplate,
  Loader2,
  Save,
  Send,
  Smartphone,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { api } from "@/lib/api";
import type { Campaign, Mode, Project } from "@/types";
const markdown =
  "# We have news.\n\nHello! 👋\n\nA new edition of our newsletter is here.\n\n## What we are building\n\n- **Something new** to make your day easier\n- Improvements you asked for\n- Behind-the-scenes stories and next steps\n\n[Explore the news →](https://hexascore.app)\n\n---\n\nThanks for being part of this journey.\n\nUntil next time,\n\n**Alex**";
const mdx =
  '# An idea brought to life.\n\nHello! This edition combines **Markdown** with React Email components.\n\n<Section style={{ backgroundColor: "#f5f5f5", padding: "24px", borderRadius: "8px" }}>\n  <Heading as="h2">The next chapter starts now.</Heading>\n  <Text>Updates worth a place in your inbox.</Text>\n  <Button href="https://hexascore.app" style={{ backgroundColor: "#171717", color: "#ffffff", padding: "12px 20px", borderRadius: "6px" }}>Explore the project →</Button>\n</Section>\n\nThanks for following along!';
const templates = {
  welcome:
    '<Section>\n  <Heading>Welcome aboard.</Heading>\n  <Text>You are now part of our community. Every update is a chance to build something better.</Text>\n  <Button href="https://hexascore.app" style={{ backgroundColor: "#171717", color: "#ffffff", padding: "12px 20px", borderRadius: "6px" }}>Explore the project →</Button>\n  <Hr />\n  <Text>Until next time, Alex.</Text>\n</Section>',
  update:
    '<Section>\n  <Text>PROJECT UPDATES</Text>\n  <Heading>A new release. New possibilities.</Heading>\n  <Text>We are excited to share what we have just launched.</Text>\n  <Row>\n    <Column><Heading as="h3">Simpler</Heading><Text>An experience designed down to the details.</Text></Column>\n    <Column><Heading as="h3">Faster</Heading><Text>Less waiting. More time for what matters.</Text></Column>\n  </Row>\n  <Button href="https://hexascore.app">See what changed →</Button>\n</Section>',
  digest:
    '<Section>\n  <Text>YOUR ROUNDUP</Text>\n  <Heading>What deserves your attention.</Heading>\n  <Heading as="h2">01. A discovery</Heading>\n  <Text>An interesting idea to start the week.</Text>\n  <Hr />\n  <Heading as="h2">02. Behind the scenes</Heading>\n  <Text>A glimpse of what we are building.</Text>\n  <Hr />\n  <Link href="https://hexascore.app">Continue the conversation →</Link>\n</Section>',
};
const snippets: Record<string, string> = {
  Heading: "<Heading>Your heading here</Heading>",
  Text: "<Text>Write your message here.</Text>",
  Button:
    '<Button href="https://example.com" style={{ backgroundColor: "#171717", color: "#ffffff", padding: "12px 20px", borderRadius: "6px" }}>Learn more →</Button>',
  Section:
    '<Section style={{ padding: "24px" }}>\n  <Text>Your content here.</Text>\n</Section>',
  Row: "<Row>\n  <Column><Text>Column 1</Text></Column>\n  <Column><Text>Column 2</Text></Column>\n</Row>",
  Img: '<Img src="https://example.com/image.png" alt="Image description" width="480" />',
  Hr: "<Hr />",
  Link: '<Link href="https://example.com">Open link →</Link>',
  Container: "<Container>\n  <Text>Your content here.</Text>\n</Container>",
  CodeInline: '<CodeInline>const hello = "world"</CodeInline>',
  CodeBlock: '<CodeBlock code={"const hello = 42;"} language="javascript" />',
  Tailwind:
    '<Tailwind><Section className="bg-gray-100 p-6 rounded-lg"><Text className="text-gray-900">Your message here.</Text></Section></Tailwind>',
  Markdown: '<Markdown>{"**Markdown** in a component."}</Markdown>',
};
export default function Composer({
  project,
  campaign,
  confirmed,
}: {
  project: Project;
  campaign?: Campaign;
  confirmed: number;
}) {
  const [mode, setMode] = useState<Mode>(campaign?.mode || "markdown");
  const [buffers, setBuffers] = useState<Record<Mode, string>>({
    markdown: campaign?.mode === "markdown" ? campaign.source : markdown,
    mdx: campaign?.mode === "mdx" ? campaign.source : mdx,
    components:
      campaign?.mode === "components" ? campaign.source : templates.welcome,
  });
  const source = buffers[mode];
  const setSource = (s: string) => setBuffers((b) => ({ ...b, [mode]: s }));
  const [subject, setSubject] = useState(campaign?.subject || ""),
    [savedId, setSavedId] = useState(campaign?.id),
    [busy, setBusy] = useState(""),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [html, setHtml] = useState(""),
    [mobile, setMobile] = useState(false),
    [confirm, setConfirm] = useState(false);
  const editor = useRef<HTMLTextAreaElement>(null);
  const data = () => ({
    project_id: String(project.id),
    id: savedId ? String(savedId) : "",
    subject,
    mode,
    source,
  });
  async function save() {
    const r = await api<Campaign>("/api/campaigns", data());
    setSavedId(r.id);
    setNotice("Draft saved.");
    return r.id;
  }
  async function act(action: "save" | "preview" | "prepare" | "send") {
    setBusy(action);
    setError("");
    setNotice("");
    try {
      if (action === "save") await save();
      if (action === "preview") {
        const r = await api<{ url: string }>("/api/preview", data());
        setHtml(r.url);
      }
      if (action === "prepare") {
        await save();
        setConfirm(true);
      }
      if (action === "send") {
        const r = await api(`/api/campaigns/${savedId}/send`, {});
        setConfirm(false);
        router.visit(`/campaigns?project=${project.id}`);
      }
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy("");
    }
  }
  function insert(snippet: string) {
    const el = editor.current;
    const pos = el?.selectionStart ?? source.length;
    setSource(
      source.slice(0, pos) +
        "\n\n" +
        snippet +
        "\n" +
        source.slice(el?.selectionEnd ?? pos),
    );
    requestAnimationFrame(() => el?.focus());
  }
  return (
    <>
      <div className="mb-7 flex flex-wrap gap-4 justify-between items-start">
        <div>
          <button
            className="text-xs text-muted-foreground flex gap-1.5 items-center mb-4 hover:text-foreground"
            onClick={() => router.visit(`/campaigns?project=${project.id}`)}
          >
            <ArrowLeft size={13} />
            All campaigns
          </button>
          <h1 className="page-title">
            {campaign ? "Edit campaign" : "A new conversation."}
          </h1>
          <p className="page-description">
            Write something people want to open.
          </p>
        </div>
        <div className="flex gap-2 mt-6">
          <Button
            variant="outline"
            disabled={!!busy}
            onClick={() => act("save")}
          >
            <Save size={14} />
            Save draft
          </Button>
          <Button
            disabled={!!busy || confirmed === 0}
            onClick={() => act("prepare")}
          >
            <Send size={14} />
            Review send
          </Button>
        </div>
      </div>
      {error && (
        <div role="alert" className="notice error mb-5">
          {error}
        </div>
      )}
      {notice && (
        <div role="status" className="notice mb-5">
          <Check size={14} />
          {notice}
        </div>
      )}
      <div className="grid grid-cols-1 xl:grid-cols-[1fr_1fr] gap-5">
        <section className="panel overflow-hidden">
          <div className="p-5 border-b border-border">
            <label
              className="text-xs text-muted-foreground block mb-2"
              htmlFor="subject"
            >
              Email subject
            </label>
            <Input
              id="subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              maxLength={200}
              placeholder="A good story starts with a good subject."
              className="bg-transparent border-0 px-0 shadow-none text-base focus-visible:ring-0"
            />
            <div className="text-[11px] text-muted-foreground mt-3">
              From: <span className="text-foreground">{project.from_email}</span>
            </div>
          </div>
          <div className="p-3 flex gap-1 border-b border-border">
            {(
              [
                { mode: "markdown", label: "Markdown", icon: FileText },
                { mode: "mdx", label: "MDX", icon: Code2 },
                {
                  mode: "components",
                  label: "React Email",
                  icon: LayoutTemplate,
                },
              ] as const
            ).map((item) => (
              <Button
                key={item.mode}
                size="sm"
                variant={mode === item.mode ? "secondary" : "ghost"}
                onClick={() => setMode(item.mode)}
              >
                <item.icon size={13} />
                {item.label}
              </Button>
            ))}
          </div>
          {mode === "components" && (
            <div className="p-3 border-b border-border space-y-3">
              <div className="flex gap-2 items-center">
                <span className="text-[10px] text-muted-foreground uppercase tracking-wider">
                  Templates
                </span>
                {Object.entries({
                  welcome: "Welcome",
                  update: "Launch",
                  digest: "Digest",
                }).map(([key, label]) => (
                  <Button
                    key={key}
                    variant="outline"
                    size="sm"
                    className="h-7 text-[10px]"
                    onClick={() =>
                      setSource(templates[key as keyof typeof templates])
                    }
                  >
                    {label}
                  </Button>
                ))}
              </div>
              <div className="flex flex-wrap gap-1.5">
                {Object.keys(snippets).map((name) => (
                  <button
                    key={name}
                    onClick={() => insert(snippets[name])}
                    className="text-[10px] border border-border bg-muted/50 hover:bg-muted px-2 py-1 rounded"
                  >
                    + {name}
                  </button>
                ))}
              </div>
            </div>
          )}
          <Textarea
            ref={editor}
            aria-label="Campaign content"
            value={source}
            onChange={(e) => setSource(e.target.value)}
            spellCheck={false}
            className="min-h-[510px] rounded-none border-0 shadow-none p-5 font-mono text-xs leading-6 resize-y focus-visible:ring-0 bg-transparent"
          />
          <div className="px-5 py-3 border-t border-border text-[10px] text-muted-foreground flex justify-between">
            <span>
              {mode === "markdown"
                ? "Markdown with tables, lists, and code blocks"
                : "React Email components · isolated MDX"}
            </span>
            <span>{source.length.toLocaleString("en-US")} characters</span>
          </div>
        </section>
        <section className="panel overflow-hidden flex flex-col">
          <div className="px-5 py-4 border-b border-border flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-medium">
              <Eye size={14} />
              Email preview
            </div>
            <div className="flex gap-2">
              <Button
                variant="ghost"
                size="icon"
                aria-label="Toggle mobile preview"
                onClick={() => setMobile(!mobile)}
                className={mobile ? "bg-muted" : ""}
              >
                <Smartphone size={14} />
              </Button>
              <Button
                size="sm"
                variant="outline"
                disabled={!!busy}
                onClick={() => act("preview")}
              >
                {busy === "preview" ? (
                  <Loader2 className="animate-spin" size={13} />
                ) : (
                  <Eye size={13} />
                )}
                Refresh preview
              </Button>
            </div>
          </div>
          <div className="flex-1 bg-[#ededed] min-h-[560px] p-4 flex justify-center">
            {html ? (
              <iframe
                title="Campaign preview"
                sandbox=""
                src={html}
                className={`bg-white border-0 w-full min-h-[560px] ${mobile ? "max-w-[375px]" : "max-w-full"}`}
              />
            ) : (
              <div className="self-center text-center max-w-xs text-neutral-500">
                <div className="size-14 mx-auto rounded-xl bg-white grid place-items-center mb-5 border border-neutral-200">
                  <MailPreview />
                </div>
                <h3 className="font-medium text-sm text-neutral-800">
                  Your message, on the other side.
                </h3>
                <p className="text-xs leading-6 mt-2">
                  Write your content and refresh the preview to see how the
                  email will arrive.
                </p>
              </div>
            )}
          </div>
        </section>
      </div>
      <div className="mt-5 text-[11px] text-muted-foreground flex gap-2 items-center">
        <span className="size-1.5 rounded-full bg-emerald-400" />
        An unsubscribe footer is included automatically in every email.
      </div>
      <Dialog open={confirm} onOpenChange={setConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ready for the inbox?</DialogTitle>
            <DialogDescription>
              Sending starts when you confirm. Review the subject, content, and
              project before continuing.
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-lg bg-muted/40 border border-border p-4 space-y-2 text-sm">
            <p>{subject}</p>
            <p className="text-muted-foreground text-xs">
              Project: {project.name}
            </p>
            <p className="text-muted-foreground text-xs">
              Up to {confirmed} confirmed subscribers · {project.from_email}
            </p>
          </div>
          <p className="text-xs text-muted-foreground">
            Only confirmed subscriptions are eligible. Unsubscribe status is
            checked again before delivery.
          </p>
          <Button disabled={!!busy} onClick={() => act("send")}>
            {busy === "send"
              ? "Queuing…"
              : "Confirm and send campaign"}
            <Send size={14} />
          </Button>
        </DialogContent>
      </Dialog>
    </>
  );
}
function MailPreview() {
  return <Send size={23} className="text-neutral-600" />;
}
