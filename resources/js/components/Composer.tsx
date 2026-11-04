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
  "# Temos novidades.\n\nOlá! 👋\n\nUma nova edição da nossa newsletter acaba de chegar.\n\n## O que estamos construindo\n\n- **Algo novo** para tornar seu dia mais simples\n- Melhorias que você pediu\n- Bastidores e próximos passos\n\n[Conheça as novidades →](https://hexascore.app)\n\n---\n\nObrigado por fazer parte dessa jornada.\n\nAté a próxima,\n\n**Alex**";
const mdx =
  '# Uma ideia que virou realidade.\n\nOlá! Esta edição mistura **Markdown** com componentes React Email.\n\n<Section style={{ backgroundColor: "#f5f5f5", padding: "24px", borderRadius: "8px" }}>\n  <Heading as="h2">O próximo capítulo começa agora.</Heading>\n  <Text>Novidades que merecem um lugar na sua caixa de entrada.</Text>\n  <Button href="https://hexascore.app" style={{ backgroundColor: "#171717", color: "#ffffff", padding: "12px 20px", borderRadius: "6px" }}>Conhecer o projeto →</Button>\n</Section>\n\nObrigado por acompanhar!';
const templates = {
  welcome:
    '<Section>\n  <Heading>Que bom ter você por aqui.</Heading>\n  <Text>Você agora faz parte da nossa comunidade. Aqui, cada novidade é uma oportunidade de construir algo melhor.</Text>\n  <Button href="https://hexascore.app" style={{ backgroundColor: "#171717", color: "#ffffff", padding: "12px 20px", borderRadius: "6px" }}>Conheça o projeto →</Button>\n  <Hr />\n  <Text>Até a próxima, Alex.</Text>\n</Section>',
  update:
    '<Section>\n  <Text>NOVIDADES DO PROJETO</Text>\n  <Heading>Uma nova versão. Novas possibilidades.</Heading>\n  <Text>Estamos animados para compartilhar o que acabamos de lançar.</Text>\n  <Row>\n    <Column><Heading as="h3">Mais simples</Heading><Text>Uma experiência pensada nos detalhes.</Text></Column>\n    <Column><Heading as="h3">Mais rápido</Heading><Text>Menos espera. Mais tempo para o que importa.</Text></Column>\n  </Row>\n  <Button href="https://hexascore.app">Ver o que mudou →</Button>\n</Section>',
  digest:
    '<Section>\n  <Text>SUA DOSE DE NOVIDADES</Text>\n  <Heading>O que vale a sua atenção.</Heading>\n  <Heading as="h2">01. Uma descoberta</Heading>\n  <Text>Uma ideia interessante para começar a semana.</Text>\n  <Hr />\n  <Heading as="h2">02. Nos bastidores</Heading>\n  <Text>Um pouco do que estamos construindo.</Text>\n  <Hr />\n  <Link href="https://hexascore.app">Continue a conversa →</Link>\n</Section>',
};
const snippets: Record<string, string> = {
  Heading: "<Heading>Seu título aqui</Heading>",
  Text: "<Text>Escreva sua mensagem aqui.</Text>",
  Button:
    '<Button href="https://example.com" style={{ backgroundColor: "#171717", color: "#ffffff", padding: "12px 20px", borderRadius: "6px" }}>Saiba mais →</Button>',
  Section:
    '<Section style={{ padding: "24px" }}>\n  <Text>Seu conteúdo aqui.</Text>\n</Section>',
  Row: "<Row>\n  <Column><Text>Coluna 1</Text></Column>\n  <Column><Text>Coluna 2</Text></Column>\n</Row>",
  Img: '<Img src="https://example.com/image.png" alt="Descrição da imagem" width="480" />',
  Hr: "<Hr />",
  Link: '<Link href="https://example.com">Abrir link →</Link>',
  Container: "<Container>\n  <Text>Seu conteúdo aqui.</Text>\n</Container>",
  CodeInline: '<CodeInline>const hello = "world"</CodeInline>',
  CodeBlock: '<CodeBlock code={"const hello = 42;"} language="javascript" />',
  Tailwind:
    '<Tailwind><Section className="bg-gray-100 p-6 rounded-lg"><Text className="text-gray-900">Sua mensagem aqui.</Text></Section></Tailwind>',
  Markdown: '<Markdown>{"**Markdown** em um componente."}</Markdown>',
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
    setNotice("Rascunho salvo.");
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
            Todas as campanhas
          </button>
          <h1 className="page-title">
            {campaign ? "Editar campanha" : "Uma nova conversa."}
          </h1>
          <p className="page-description">
            Escreva algo que as pessoas queiram abrir.
          </p>
        </div>
        <div className="flex gap-2 mt-6">
          <Button
            variant="outline"
            disabled={!!busy}
            onClick={() => act("save")}
          >
            <Save size={14} />
            Salvar rascunho
          </Button>
          <Button
            disabled={!!busy || confirmed === 0}
            onClick={() => act("prepare")}
          >
            <Send size={14} />
            Revisar envio
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
              Assunto do e-mail
            </label>
            <Input
              id="subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              maxLength={200}
              placeholder="Uma boa história começa com um bom assunto."
              className="bg-transparent border-0 px-0 shadow-none text-base focus-visible:ring-0"
            />
            <div className="text-[11px] text-muted-foreground mt-3">
              De: <span className="text-foreground">{project.from_email}</span>
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
                  Modelos
                </span>
                {Object.entries({
                  welcome: "Boas-vindas",
                  update: "Lançamento",
                  digest: "Resumo",
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
            aria-label="Conteúdo da campanha"
            value={source}
            onChange={(e) => setSource(e.target.value)}
            spellCheck={false}
            className="min-h-[510px] rounded-none border-0 shadow-none p-5 font-mono text-xs leading-6 resize-y focus-visible:ring-0 bg-transparent"
          />
          <div className="px-5 py-3 border-t border-border text-[10px] text-muted-foreground flex justify-between">
            <span>
              {mode === "markdown"
                ? "Markdown com tabelas, listas e blocos de código"
                : "Componentes React Email · MDX isolado"}
            </span>
            <span>{source.length.toLocaleString("pt-BR")} caracteres</span>
          </div>
        </section>
        <section className="panel overflow-hidden flex flex-col">
          <div className="px-5 py-4 border-b border-border flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-medium">
              <Eye size={14} />
              Prévia do e-mail
            </div>
            <div className="flex gap-2">
              <Button
                variant="ghost"
                size="icon"
                aria-label="Alternar prévia mobile"
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
                Atualizar prévia
              </Button>
            </div>
          </div>
          <div className="flex-1 bg-[#ededed] min-h-[560px] p-4 flex justify-center">
            {html ? (
              <iframe
                title="Prévia da campanha"
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
                  Sua mensagem, do outro lado.
                </h3>
                <p className="text-xs leading-6 mt-2">
                  Escreva seu conteúdo e atualize a prévia para ver como o
                  e-mail vai chegar.
                </p>
              </div>
            )}
          </div>
        </section>
      </div>
      <div className="mt-5 text-[11px] text-muted-foreground flex gap-2 items-center">
        <span className="size-1.5 rounded-full bg-emerald-400" />O rodapé de
        descadastro é incluído automaticamente em todos os e-mails.
      </div>
      <Dialog open={confirm} onOpenChange={setConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Pronto para chegar à caixa de entrada?</DialogTitle>
            <DialogDescription>
              O envio começa assim que você confirmar. Revise o assunto, o
              conteúdo e o projeto antes de continuar.
            </DialogDescription>
          </DialogHeader>
          <div className="rounded-lg bg-muted/40 border border-border p-4 space-y-2 text-sm">
            <p>{subject}</p>
            <p className="text-muted-foreground text-xs">
              Projeto: {project.name}
            </p>
            <p className="text-muted-foreground text-xs">
              Até {confirmed} inscritos confirmados · {project.from_email}
            </p>
          </div>
          <p className="text-xs text-muted-foreground">
            Somente inscrições confirmadas serão consideradas. Descadastros são
            verificados novamente no momento do envio.
          </p>
          <Button disabled={!!busy} onClick={() => act("send")}>
            {busy === "send"
              ? "Colocando na fila…"
              : "Confirmar e enviar campanha"}
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
