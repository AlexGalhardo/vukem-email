import { Link, router, usePage } from "@inertiajs/react";
import {
  ArrowUpRight,
  ChevronDown,
  CircleHelp,
  FileText,
  Globe2,
  LayoutDashboard,
  LogOut,
  Mail,
  Plus,
  Send,
  Settings2,
  Users,
} from "lucide-react";
import type { ReactNode } from "react";
import type { Project } from "@/types";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/subscribers", label: "Subscribers", icon: Users },
  { href: "/campaigns", label: "Campaigns", icon: Mail },
  { href: "/projects", label: "Projects", icon: Globe2 },
];
export default function AppLayout({
  children,
  project,
  projects,
}: {
  children: ReactNode;
  project: Project;
  projects: Project[];
}) {
  const path = usePage().url.split("?")[0];
  const link = (href: string) =>
    `${href}${project.id ? `?project=${project.id}` : ""}`;
  return (
    <SidebarProvider>
      <Sidebar className="border-r border-border">
        <SidebarHeader className="p-5">
          <Link
            href={link("/")}
            className="flex items-center gap-2.5 font-semibold tracking-tight text-lg"
          >
            <span className="brand-mark">
              <Send size={17} />
            </span>
            vukem
            <span className="font-normal text-muted-foreground">email</span>
          </Link>
          <div className="mt-7 mb-2 text-[10px] tracking-[.14em] text-muted-foreground font-medium">
            WORKSPACE
          </div>
          <Select
            value={project.id ? String(project.id) : undefined}
            onValueChange={(value) => router.get(path, { project: value })}
          >
            <SelectTrigger className="w-full bg-sidebar-accent border-border h-11">
              <div className="flex items-center gap-2">
                <span className="flex size-6 items-center justify-center rounded bg-white/10 text-xs font-semibold">
                  {project.name?.slice(0, 1) || "+"}
                </span>
                <SelectValue placeholder="Select a project" />
              </div>
            </SelectTrigger>
            <SelectContent>
              {projects.map((p) => (
                <SelectItem key={p.id} value={String(p.id)}>
                  {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup className="px-3">
            <SidebarGroupLabel className="text-[10px] tracking-[.14em] px-3 mb-2">
              MANAGE
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {nav.map((item) => (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton
                      asChild
                      isActive={
                        path === item.href ||
                        (item.href === "/campaigns" &&
                          path.startsWith("/campaigns"))
                      }
                      className="h-10 px-3 rounded-md"
                    >
                      <Link href={link(item.href)}>
                        <item.icon size={16} />
                        <span>{item.label}</span>
                        {item.href === "/projects" && (
                          <span className="ml-auto text-xs text-muted-foreground">
                            {projects.length}
                          </span>
                        )}
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
          <SidebarGroup className="px-3 mt-4">
            <SidebarGroupLabel className="text-[10px] tracking-[.14em] px-3 mb-2">
              WORKSPACE
            </SidebarGroupLabel>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={path === "/settings"}
                  className="h-10 px-3"
                >
                  <Link href={link("/settings")}>
                    <Settings2 size={16} />
                    Settings
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton asChild className="h-10 px-3">
                  <a
                    href="https://email.vukem.dev/docs/introduction"
                    target="_blank"
                    rel="noreferrer"
                  >
                    <FileText size={16} />
                    Documentation
                    <ArrowUpRight className="ml-auto" size={13} />
                  </a>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter className="p-4">
          <div className="rounded-lg border border-border p-3 mb-3">
            <div className="flex gap-2 items-center text-xs font-medium">
              <span className="size-1.5 bg-emerald-400 rounded-full" />
              Made for your projects
            </div>
            <p className="text-xs leading-5 text-muted-foreground mt-2">
              Your own audience.
              <br />
              One connection at a time.
            </p>
          </div>
          <div className="flex items-center gap-2.5 px-1">
            <div className="rounded-full bg-[#c6efb9] text-[#20451a] size-8 grid place-items-center text-xs font-semibold">
              AG
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-xs font-medium">Alex Galhardo</div>
              <div className="text-[10px] text-muted-foreground truncate">
                Personal workspace
              </div>
            </div>
            <button
              aria-label="Sign out"
              className="text-muted-foreground hover:text-foreground"
              onClick={() => router.post("/logout")}
            >
              <LogOut size={15} />
            </button>
          </div>
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="h-16 flex items-center border-b border-border px-5 md:px-9 gap-3">
          <SidebarTrigger className="md:hidden" />
          <span className="text-muted-foreground text-sm">
            Personal workspace
          </span>
          <span className="text-muted-foreground/50">/</span>
          <span className="text-sm">{project.name || "Your projects"}</span>
          <span className="ml-1 text-[10px] rounded border border-border px-1.5 py-0.5 text-muted-foreground">
            Personal
          </span>
          <div className="ml-auto flex gap-4 items-center">
            <a
              href="https://email.vukem.dev/docs/introduction"
              target="_blank"
              rel="noreferrer"
              aria-label="Help"
              className="text-muted-foreground"
            >
              <CircleHelp size={17} />
            </a>
            <div className="size-7 rounded-full border border-border grid place-items-center text-[10px] bg-muted">
              AG
            </div>
          </div>
        </header>
        <main className="mx-auto w-full max-w-[1500px] px-5 py-8 md:px-9 md:py-10">
          {children}
        </main>
        <footer className="mt-auto mx-5 md:mx-9 py-5 border-t border-border flex justify-between text-[11px] text-muted-foreground">
          <span>© {new Date().getFullYear()} Vukem Email</span>
          <span className="flex gap-1.5 items-center">
            <span className="size-1.5 rounded-full bg-emerald-400" />
            Your next connection starts here
          </span>
        </footer>
      </SidebarInset>
    </SidebarProvider>
  );
}
