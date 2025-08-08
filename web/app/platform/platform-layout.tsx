"use client";

import { 
  Users, 
  Shield, 
  Clock,
  ArrowLeft,
  Settings
} from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { type Session } from "next-auth";

import { KloudliteLogo } from "@/components/kloudlite-logo";
import { NotificationBell } from "@/components/notification-bell";
import { SidebarNav } from "@/components/platform/sidebar-nav";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { UserMenu } from "@/components/user-menu";

interface PlatformLayoutProps {
  platformRole: {
    role: string;
    canCreateTeams: boolean;
    canManagePlatform: boolean;
  };
  session: Session;
  teamRequestsCount: number;
  children: React.ReactNode;
}

export function PlatformLayout({
  platformRole,
  session,
  teamRequestsCount,
  children,
}: PlatformLayoutProps) {
  const pathname = usePathname();
  const router = useRouter();
  const isSuperAdmin = platformRole.role === "super_admin";

  // Determine active page from pathname
  const activeTab = pathname === "/platform" ? "overview" :
                   pathname === "/platform/requests" ? "requests" :
                   pathname === "/platform/users" ? "users" :
                   pathname === "/platform/providers" ? "providers" :
                   pathname === "/platform/settings" ? "settings" : "overview";

  // Navigation items
  const navItems = [
    { id: "overview", label: "Overview", icon: Shield, href: "/platform" },
    { id: "requests", label: "Team Requests", icon: Clock, href: "/platform/requests", badge: teamRequestsCount },
    ...(isSuperAdmin ? [
      { id: "users", label: "Users", icon: Users, href: "/platform/users" },
      { id: "settings", label: "Settings", icon: Settings, href: "/platform/settings" }
    ] : [])
  ];

  const handleNavClick = (id: string) => {
    const item = navItems.find(item => item.id === id);
    if (item) {
      router.push(item.href);
    }
  };

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar Navigation */}
      <aside className="w-[260px] border-r bg-card/30 flex flex-col">
        <div className="p-4">
          <div className="flex items-center gap-2.5 rounded-lg bg-muted/30 p-3">
            <div className="h-9 w-9 rounded-md bg-gradient-to-br from-primary/20 to-primary/10 flex items-center justify-center">
              <Shield className="h-4 w-4 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-[13px]">Platform</h3>
              <p className="text-[11px] text-muted-foreground capitalize">{platformRole.role}</p>
            </div>
          </div>
        </div>
        
        {/* Scrollable navigation area */}
        <ScrollArea className="flex-1 px-3 py-4">
          <div className="pb-6">
            <h2 className="mb-2.5 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">
              Management
            </h2>
            <SidebarNav 
              items={navItems.map(item => ({
                id: item.id,
                label: item.label,
                icon: item.icon,
              }))}
              activeItem={activeTab}
              onItemClick={handleNavClick}
            />
          </div>
        </ScrollArea>
        
        {/* Theme toggle at bottom */}
        <div className="border-t p-3">
          <div className="flex items-center justify-center">
            <ThemeToggle />
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="border-b bg-background shadow-sm">
          <div className="container max-w-5xl mx-auto px-6">
            <div className="flex h-16 items-center">
              {/* Logo on the left */}
              <div className="flex items-center gap-3">
                <KloudliteLogo className="h-6" />
                <Separator orientation="vertical" className="h-6" />
                <span className="text-muted-foreground">/</span>
                <span className="font-medium">Platform</span>
              </div>
              <div className="ml-auto flex items-center gap-3">
                <NotificationBell />
                <UserMenu user={session.user} canManagePlatform={true} />
              </div>
            </div>
          </div>
        </header>
        
        {/* Content */}
        <main className="flex-1 overflow-hidden">
          <ScrollArea className="h-full">
            <div className="container max-w-5xl mx-auto p-6 lg:p-8">
              {children}
            </div>
          </ScrollArea>
        </main>
      </div>
    </div>
  );
}