import { Head, Link } from "@inertiajs/react";
import { ArrowLeft, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
export default function ErrorPage({
  status,
  message,
}: {
  status: number;
  message: string;
}) {
  return (
    <>
      <Head title={`Erro ${status}`} />
      <main className="auth-page">
        <section className="text-center max-w-md">
          <span className="brand-mark mx-auto mb-8">
            <Send size={17} />
          </span>
          <p className="eyebrow mb-3">ERRO {status}</p>
          <h1 className="text-2xl font-semibold tracking-tight">
            {status === 403
              ? "Este acesso não está disponível."
              : status === 404
                ? "Não encontramos esta página."
                : status >= 500
                  ? "Algo não saiu como esperado."
                  : "Não foi possível continuar."}
          </h1>
          <p className="text-sm text-muted-foreground mt-4 leading-6">
            {message}
          </p>
          <Button asChild variant="outline" className="mt-7">
            <Link href="/">
              <ArrowLeft size={14} />
              Voltar ao início
            </Link>
          </Button>
        </section>
      </main>
    </>
  );
}
