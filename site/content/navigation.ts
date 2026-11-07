export type NavCategory = { title: string; dir: string; prefix?: string; pages: string[] };
export const navigation: NavCategory[] = [
  { title: "Getting started", dir: "getting-started", prefix: "", pages: ["introduction", "installation", "configuration", "deployment"] },
  { title: "Concepts", dir: "concepts", pages: ["projects", "subscribers", "campaigns", "delivery"] },
  { title: "Guides", dir: "guides", pages: ["google-oidc", "email-content", "subscribe-api", "operations", "security"] },
];
