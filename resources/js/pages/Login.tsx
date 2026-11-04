import { Head } from "@inertiajs/react";
import { ArrowUpRight, LockKeyhole, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
export default function Login({ configured }: { configured: boolean }) {
  return (
    <>
      <Head title="Sign in" />
      <main className="auth-page">
        <a
          href="/login"
          className="absolute top-8 left-8 flex items-center gap-2 font-semibold"
        >
          <span className="brand-mark">
            <Send size={17} />
          </span>
          vukem email
        </a>
        <section className="w-full max-w-sm text-center relative">
          <div className="mx-auto mb-7 rounded-2xl border border-border bg-card size-16 grid place-items-center">
            <MailIcon />
          </div>
          <p className="eyebrow mb-4">YOUR NEWSLETTER WORKSPACE</p>
          <h1 className="text-3xl font-semibold tracking-tight">
            Good ideas deserve
            <br />
            to go further.
          </h1>
          <p className="mt-4 text-sm text-muted-foreground leading-6">
            Sign in to create campaigns, follow your audience, and connect
            with your subscribers.
          </p>
          <Button
            asChild
            className="w-full mt-8 h-11 bg-white text-black hover:bg-white/90"
          >
            <a href="/auth/google/redirect">
              <svg
                width="17"
                height="17"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  fill="#4285F4"
                  d="M22.6 12.2c0-.7-.1-1.4-.2-2.1H12v4h5.9a5.1 5.1 0 0 1-2.2 3.4v2.8h3.6c2.1-1.9 3.3-4.7 3.3-8.1"
                />
                <path
                  fill="#34A853"
                  d="M12 23c3 0 5.5-1 7.3-2.7l-3.6-2.8c-1 .6-2.2 1-3.7 1-2.8 0-5.2-1.9-6.1-4.4H2.2V17A11 11 0 0 0 12 23"
                />
                <path
                  fill="#FBBC05"
                  d="M5.9 14.1a6.6 6.6 0 0 1 0-4.2V7H2.2a11 11 0 0 0 0 10z"
                />
                <path
                  fill="#EA4335"
                  d="M12 5.5c1.6 0 3 .5 4.1 1.6l3.1-3.2A10.5 10.5 0 0 0 12 1a11 11 0 0 0-9.8 6l3.7 2.9C6.8 7.4 9.2 5.5 12 5.5"
                />
              </svg>
              Continue with Google
            </a>
          </Button>
          <div className="text-xs text-muted-foreground flex gap-1.5 items-center justify-center mt-5">
            <LockKeyhole size={12} />
            Administrator access only
          </div>
          {!configured && (
            <p
              role="status"
              className="mt-6 text-xs text-amber-300 border border-amber-300/20 rounded-lg p-3"
            >
              Sign-in will be available after Google OAuth credentials are
              configured.
            </p>
          )}
        </section>
        <div className="absolute bottom-8 text-xs text-muted-foreground">
          Built with Vukem <span className="mx-2">·</span> Made to connect.
        </div>
      </main>
    </>
  );
}
function MailIcon() {
  return <Send className="text-foreground" size={26} />;
}
