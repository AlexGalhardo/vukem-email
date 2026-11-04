export type NavCategory = { title: string; dir: string; prefix?: string; pages: string[] };
export const navigation: NavCategory[] = [
  { title: "Comece aqui", dir: "getting-started", prefix: "", pages: ["introduction", "installation", "configuration", "deployment"] },
  { title: "Conceitos", dir: "concepts", pages: ["projects", "subscribers", "campaigns", "delivery"] },
  { title: "Guias", dir: "guides", pages: ["google-oidc", "email-content", "operations", "security"] },
];
