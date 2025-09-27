"use client";

import { 
  Cloud, 
  Box, 
  Database,
  Users,
  Settings,
  ArrowLeft,
  Home,
  Shield,
  ChevronsUpDown,
  Plus,
  Menu,
  X
} from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { type Session } from "next-auth";
import { useState, useEffect } from "react";

import { KloudliteLogo } from "@/components/kloudlite-logo";
import { NotificationBell } from "@/components/notification-bell";
import { SidebarNav } from "@/components/platform/sidebar-nav";
import { ThemeTogglePill } from "@/components/theme-toggle-variants";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { UserMenu } from "@/components/user-menu";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface TeamLayoutProps {
  team: any;
  userRole: string;
  session: Session;
  teamSlug: string;
  canManagePlatform?: boolean;
  children: React.ReactNode;
}

export function TeamLayout({
  team,
  userRole,
  session,
  teamSlug,
  canManagePlatform = false,
  children,
}: TeamLayoutProps) {
  const pathname = usePathname();
  const router = useRouter();
  const isAdmin = userRole === "owner" || userRole === "admin";
  const [teams, setTeams] = useState<any[]>([]);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Close sidebar when route changes on mobile
  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname]);

  // Fetch user teams for the selector
  useEffect(() => {
    fetch('/api/teams')
      .then(res => res.json())
      .then(data => setTeams(data.teams || []))
      .catch(() => setTeams([]));
  }, []);

  // Determine active page from pathname
  const activeTab = pathname === `/${teamSlug}` ? "overview" :
                   pathname === `/${teamSlug}/environments` ? "environments" :
                   pathname === `/${teamSlug}/workspaces` ? "workspaces" :
                   pathname === `/${teamSlug}/services` ? "services" :
                   pathname === `/${teamSlug}/management` ? "management" :
                   pathname === `/${teamSlug}/settings` ? "settings" : "overview";

  // Navigation items
  const navItems = [
    { 
      id: "overview", 
      label: "Overview", 
      icon: Home, 
      href: `/${teamSlug}`
    },
    { 
      id: "environments", 
      label: "Environments", 
      icon: Cloud, 
      href: `/${teamSlug}/environments`
    },
    { 
      id: "workspaces", 
      label: "Workspaces", 
      icon: Box, 
      href: `/${teamSlug}/workspaces`
    },
    { 
      id: "services", 
      label: "Services", 
      icon: Database, 
      href: `/${teamSlug}/services`
    },
    ...(isAdmin ? [
      { 
        id: "management", 
        label: "Members", 
        icon: Users, 
        href: `/${teamSlug}/management`
      },
      { 
        id: "settings", 
        label: "Settings", 
        icon: Settings, 
        href: `/${teamSlug}/settings`
      }
    ] : [])
  ];

  const handleNavClick = (id: string) => {
    const item = navItems.find(item => item.id === id);
    if (item) {
      router.push(item.href);
      // Close sidebar on mobile after navigation
      setSidebarOpen(false);
    }
  };

  return (
    <div className="flex h-screen bg-background">
      {/* Mobile sidebar backdrop */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 bg-background/80 backdrop-blur-sm z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar Navigation - Hidden on mobile by default, visible on lg screens */}
      <aside className={cn(
        "fixed lg:relative inset-y-0 left-0 z-50 w-[260px] border-r bg-background lg:bg-card/30 flex flex-col transform transition-transform duration-200 ease-in-out",
        sidebarOpen ? "translate-x-0" : "-translate-x-full",
        "lg:translate-x-0"
      )}>
        <div className="p-4">
          {/* Mobile close button */}
          <div className="flex items-center justify-between mb-4 lg:hidden">
            <span className="text-sm font-medium">Navigation</span>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setSidebarOpen(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/* Team Selector */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button 
                variant="ghost" 
                className="w-full justify-between h-auto p-2.5 hover:bg-accent/50 transition-colors"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <div className="h-9 w-9 rounded-md bg-gradient-to-br from-primary/20 to-primary/10 flex items-center justify-center flex-shrink-0">
                    <Shield className="h-4 w-4 text-primary" />
                  </div>
                  <div className="flex-1 text-left min-w-0 overflow-hidden">
                    <h3 className="font-semibold text-[13px] truncate">{team?.displayName}</h3>
                    <p className="text-[11px] text-muted-foreground truncate capitalize">{userRole}</p>
                  </div>
                </div>
                <ChevronsUpDown className="h-4 w-4 text-muted-foreground flex-shrink-0" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-64" align="start" sideOffset={5}>
              <DropdownMenuItem 
                key="all-teams"
                onClick={() => {
                  router.push('/overview');
                  setSidebarOpen(false);
                }}
                className="gap-2"
              >
                <Users className="h-4 w-4" />
                All Teams
              </DropdownMenuItem>
              <DropdownMenuSeparator key="separator-1" />
              {teams.map((t: any) => (
                <DropdownMenuItem
                  key={t.accountid || t.teamId}
                  onClick={() => {
                    router.push(`/${t.slug}`);
                    setSidebarOpen(false);
                  }}
                  className="gap-2"
                >
                  <Shield className="h-4 w-4" />
                  {t.displayName}
                  {t.slug === teamSlug && <span className="ml-auto text-xs">Current</span>}
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator key="separator-2" />
              <DropdownMenuItem 
                key="create-team"
                onClick={() => {
                  router.push('/teams/new');
                  setSidebarOpen(false);
                }}
                className="gap-2"
              >
                <Plus className="h-4 w-4" />
                Create New Team
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        
        {/* Scrollable navigation area */}
        <ScrollArea className="flex-1 px-2 py-4">
          <div className="space-y-7 px-1">
            {/* Main Navigation */}
            <div>
              <h2 className="mb-2.5 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">
                Resources
              </h2>
              <SidebarNav 
                items={navItems.slice(0, 4).map(item => ({
                  id: item.id,
                  label: item.label,
                  icon: item.icon,
                }))}
                activeItem={activeTab}
                onItemClick={handleNavClick}
              />
            </div>
            
            {/* Admin Section */}
            {isAdmin && navItems.length > 4 && (
              <div>
                <h2 className="mb-2.5 px-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">
                  Administration
                </h2>
                <SidebarNav 
                  items={navItems.slice(4).map(item => ({
                    id: item.id,
                    label: item.label,
                    icon: item.icon,
                  }))}
                  activeItem={activeTab}
                  onItemClick={handleNavClick}
                />
              </div>
            )}
          </div>
        </ScrollArea>
        
        {/* Bottom section with user info and theme */}
        <div className="border-t p-4">
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3 min-w-0">
              <div className="h-9 w-9 rounded-md bg-muted flex items-center justify-center flex-shrink-0">
                <Users className="h-4 w-4 text-muted-foreground" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium truncate">{session.user.email?.split('@')[0]}</p>
                <p className="text-xs text-muted-foreground truncate">{session.user.email}</p>
              </div>
            </div>
            <div className="flex justify-center">
              <ThemeTogglePill />
            </div>
          </div>
        </div>
      </aside>

      {/* Main Content Area - Wider container */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-sm">
          <div className="container max-w-7xl mx-auto px-4 lg:px-6">
            <div className="flex h-14 lg:h-16 items-center">
              {/* Mobile menu button */}
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 lg:hidden mr-2"
                onClick={() => setSidebarOpen(true)}
              >
                <Menu className="h-5 w-5" />
              </Button>

              {/* Logo on the left */}
              <div className="flex items-center gap-2 lg:gap-3">
                <KloudliteLogo className="h-5 lg:h-6" />
                <Separator orientation="vertical" className="h-5 lg:h-6" />
                <span className="text-muted-foreground text-sm lg:text-base">/</span>
                <span className="font-medium text-sm lg:text-base truncate">{team?.displayName}</span>
              </div>
              <div className="ml-auto flex items-center gap-3">
                <NotificationBell />
                <UserMenu user={session.user} canManagePlatform={canManagePlatform} />
              </div>
            </div>
          </div>
        </header>
        
        {/* Content */}
        <main className="flex-1 overflow-hidden">
          <ScrollArea className="h-full">
            <div className="container max-w-7xl mx-auto p-4 md:p-6 lg:p-8">
              {children}
            </div>
          </ScrollArea>
        </main>
      </div>
    </div>
  );
}