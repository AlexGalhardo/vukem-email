export interface Project {
  id: number;
  name: string;
  slug: string;
  domain: string;
  from_email: string;
  description: string;
}
export interface Subscriber {
  id: number;
  project_id: number;
  email: string;
  status: string;
  created_at: string;
}
export interface Campaign {
  id: number;
  project_id: number;
  subject: string;
  mode: Mode;
  source: string;
  status: string;
  recipients: number;
  created_at: string;
}
export interface Bounce {
  id: number;
  email: string;
  type: "hard" | "soft" | "complaint";
  reason: string;
  created_at: string;
}
export interface ApiToken {
  id: string;
  name: string;
  abilities: string[];
  last_used_at: string | null;
}
export type Mode = "markdown" | "mdx" | "components";
export interface DashboardProps {
  projects: Project[];
  project: Project;
  subscribers: Subscriber[];
  campaigns: Campaign[];
  stats: Record<string, number>;
  deliveryStats: Record<string, Record<string, number>>;
  trackingStats: Record<string, Record<string, number>>;
  apiTokens: ApiToken[];
  bounces: Bounce[];
  total: number;
  page: number;
  filters: { q: string; status: string };
  admin: string;
}
