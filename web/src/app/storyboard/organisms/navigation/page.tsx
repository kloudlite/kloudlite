"use client";

import { SidebarNavItem, TeamSelector, UserProfileDropdown, BreadcrumbNav, TabNav } from "@/components/organisms";
import { Breadcrumb } from "@/components/atoms";
import { LayoutDashboard, Layers, Database, Terminal, Users2, Cog } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function NavigationPage() {
  const navigation = [
    { name: 'Overview', href: '/dashboard', icon: LayoutDashboard },
    { name: 'Environments', href: '/dashboard/environments', icon: Layers },
    { name: 'Shared Services', href: '/dashboard/services', icon: Database },
    { name: 'Workspaces', href: '/dashboard/workspaces', icon: Terminal },
  ];

  const team = {
    id: 1,
    name: "Engineering Team",
    slug: "engineering",
    role: "admin",
    members: 12,
    environments: 8,
    lastAccessed: "2 hours ago"
  };

  const user = {
    id: "1",
    name: "John Doe",
    email: "john.doe@example.com"
  };

  const tabs = [
    { label: 'Overview', href: '#overview' },
    { label: 'Services', href: '#services', count: 12 },
    { label: 'Apps', href: '#apps', count: 5 },
    { label: 'Settings', href: '#settings' },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Navigation Components
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Components for navigation and wayfinding.
        </p>
      </div>

      <ComponentShowcase
        title="Breadcrumb Navigation"
        description="Shows the current page location in the site hierarchy"
      >
        <div className="space-y-4">
          <Breadcrumb 
            items={[
              { label: 'Dashboard', href: '/dashboard' },
              { label: 'Environments', href: '/dashboard/environments' },
              { label: 'Production' }
            ]}
          />

          <Breadcrumb 
            items={[
              { label: 'Home', href: '/' },
              { label: 'Documentation', href: '/docs' },
              { label: 'Getting Started', href: '/docs/getting-started' },
              { label: 'Installation' }
            ]}
          />

          <Breadcrumb 
            items={[
              { label: 'Settings', href: '/settings' },
              { label: 'Team Management' }
            ]}
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Tab Navigation"
        description="Navigation tabs for switching between views"
      >
        <div className="space-y-6">
          <div>
            <p className="text-sm font-medium mb-3">Default Tabs</p>
            <TabNav
              items={tabs}
              activeHref="#overview"
            />
          </div>

          <div>
            <p className="text-sm font-medium mb-3">Pills Style</p>
            <TabNav
              items={[
                { label: 'All Items', href: '#all' },
                { label: 'Active', href: '#active', count: 8 },
                { label: 'Archived', href: '#archived', count: 3 },
              ]}
              activeHref="#active"
              variant="pills"
            />
          </div>

          <div>
            <p className="text-sm font-medium mb-3">Underline Style</p>
            <TabNav
              items={[
                { label: 'Profile', href: '#profile' },
                { label: 'Security', href: '#security' },
                { label: 'Notifications', href: '#notifications', count: 4 },
                { label: 'Billing', href: '#billing' },
              ]}
              activeHref="#profile"
              variant="underline"
            />
          </div>

          <div>
            <p className="text-sm font-medium mb-3">With Icons</p>
            <TabNav
              items={[
                { label: 'Dashboard', href: '#dashboard', icon: LayoutDashboard },
                { label: 'Services', href: '#services', icon: Database, count: 12 },
                { label: 'Users', href: '#users', icon: Users2 },
                { label: 'Settings', href: '#settings', icon: Cog },
              ]}
              activeHref="#dashboard"
              variant="pills"
              size="lg"
            />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Sidebar Navigation Items"
        description="Navigation items for sidebar menus"
      >
        <div className="bg-slate-50 dark:bg-slate-900 p-4 rounded-lg">
          <div className="space-y-1">
            {navigation.map((item, index) => (
              <SidebarNavItem
                key={item.name}
                name={item.name}
                href={item.href}
                icon={item.icon}
                isActive={index === 0}
              />
            ))}
          </div>

          <div className="mt-6 pt-6 border-t border-slate-200 dark:border-slate-800 space-y-1">
            <SidebarNavItem
              name="User Management"
              href="/dashboard/user-management"
              icon={Users2}
              isActive={false}
            />
            <SidebarNavItem
              name="Settings"
              href="/dashboard/settings"
              icon={Cog}
              isActive={false}
            />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Team Selector"
        description="Dropdown for switching between teams"
      >
        <div className="max-w-xs">
          <TeamSelector
            selectedTeam={team}
            teamColor="from-violet-500 to-purple-600"
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="User Profile Dropdown"
        description="User profile menu with actions"
      >
        <div className="max-w-xs">
          <UserProfileDropdown user={user} />
        </div>
      </ComponentShowcase>
    </div>
  );
}