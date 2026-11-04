import { createInertiaApp } from "@inertiajs/react";
createInertiaApp({
  nonce: document.querySelector<HTMLMetaElement>('meta[property="csp-nonce"]')
    ?.nonce,
  progress: { color: "#fafafa" },
  title: (title) => `${title} — Vukem Email`,
}).finally(() => document.documentElement.removeAttribute("data-hydrating"));
