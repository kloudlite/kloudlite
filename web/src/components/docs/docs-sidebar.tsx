'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { Menu, Search, Book, Code, Zap, Shield, Cloud, Database } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Link } from '@/components/ui/link'
import { cn } from '@/lib/utils'

interface NavigationItem {
  title: string
  href: string
  icon?: React.ElementType
  badge?: string
  items?: NavigationItem[]
}

const navigationItems: NavigationItem[] = [
  {
    title: 'Getting Started',
    href: '/docs/getting-started',
    icon: Book,
    items: [
      { title: 'Introduction', href: '/docs/getting-started' },
      { title: 'Quick Start', href: '/docs/getting-started/quickstart' },
      { title: 'Installation', href: '/docs/getting-started/installation' },
    ]
  },
  {
    title: 'Core Concepts',
    href: '/docs/concepts',
    icon: Zap,
    items: [
      { title: 'Environments', href: '/docs/concepts/environments' },
      { title: 'Projects', href: '/docs/concepts/projects' },
      { title: 'Workspaces', href: '/docs/concepts/workspaces' },
      { title: 'Service Intercepts', href: '/docs/concepts/intercepts' },
    ]
  },
  {
    title: 'Development',
    href: '/docs/development',
    icon: Code,
    items: [
      { title: 'Local Development', href: '/docs/development/local' },
      { title: 'Remote Development', href: '/docs/development/remote' },
      { title: 'Debugging', href: '/docs/development/debugging' },
      { title: 'Testing', href: '/docs/development/testing' },
    ]
  },
  {
    title: 'Authentication',
    href: '/docs/authentication',
    icon: Shield,
    items: [
      { title: 'OAuth Setup', href: '/docs/authentication/oauth' },
      { title: 'Providers', href: '/docs/authentication/providers' },
      { title: 'Sessions', href: '/docs/authentication/sessions' },
      { title: 'API Keys', href: '/docs/authentication/api-keys' },
    ]
  },
  {
    title: 'Deployment',
    href: '/docs/deployment',
    icon: Cloud,
    items: [
      { title: 'Cloud Deployment', href: '/docs/deployment/cloud' },
      { title: 'Self-Hosted', href: '/docs/deployment/self-hosted' },
      { title: 'CI/CD Integration', href: '/docs/deployment/cicd' },
      { title: 'Monitoring', href: '/docs/deployment/monitoring' },
    ]
  },
  {
    title: 'API Reference',
    href: '/docs/api',
    icon: Database,
    items: [
      { title: 'Authentication API', href: '/docs/api/auth' },
      { title: 'Teams API', href: '/docs/api/teams' },
      { title: 'Resources API', href: '/docs/api/resources' },
      { title: 'Webhooks', href: '/docs/api/webhooks' },
    ]
  },
]

interface DocsSidebarProps {
  isOpen?: boolean
  onClose?: () => void
}

export function DocsSidebar({ isOpen = true, onClose }: DocsSidebarProps) {
  const pathname = usePathname()

  return (
    <aside className={cn(
      "fixed inset-y-0 left-0 z-50 w-72 bg-background border-r border-border transform transition-transform duration-300 lg:relative lg:translate-x-0 lg:z-0 flex flex-col h-screen",
      isOpen ? "translate-x-0" : "-translate-x-full"
    )}>
      {/* Sidebar Header */}
      <div className="flex-shrink-0">
        <div className="flex items-center justify-between h-16 px-6 bg-background border-b border-border">
          <div className="flex items-center gap-2">
            <Link href="/" className="text-lg font-bold">
              Kloudlite Docs
            </Link>
            <span className="text-xs px-2 py-0.5 bg-primary/10 text-primary rounded-none">
              v1.0.0
            </span>
          </div>
        </div>
        
        {/* Search */}
        <div className="px-6 py-4 bg-background border-b border-border">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search docs..."
              className="pl-10 pr-10 h-9 bg-background/50 border-muted rounded-none"
            />
            <kbd className="absolute right-2 top-1/2 -translate-y-1/2 text-xs bg-muted px-1.5 py-0.5 rounded-none text-muted-foreground">
              ⌘K
            </kbd>
          </div>
        </div>
      </div>

      {/* Navigation - Scrollable */}
      <nav className="flex-1 overflow-y-auto [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-border [&::-webkit-scrollbar-thumb]:rounded-full">
        <div className="py-4">
          {navigationItems.map((section) => {
            const isSectionActive = pathname.startsWith(section.href)
            
            return (
              <div key={section.href}>
                <div className="flex items-center gap-3 px-6 py-2 text-sm font-medium text-muted-foreground">
                  {section.icon && <section.icon className="h-4 w-4" />}
                  <span>{section.title}</span>
                  {section.badge && (
                    <span className="text-xs px-2 py-0.5 bg-primary/10 text-primary rounded-none">
                      {section.badge}
                    </span>
                  )}
                </div>

                {section.items && (
                  <div className="mt-1">
                    {section.items.map((item) => {
                      const isActive = pathname === item.href
                      return (
                        <Link
                          key={item.href}
                          href={item.href}
                          onClick={() => {
                            // Close sidebar on mobile when a link is clicked
                            if (onClose && window.innerWidth < 1024) {
                              onClose()
                            }
                          }}
                          className={cn(
                            "block px-6 py-2 text-sm transition-colors",
                            isActive 
                              ? "text-primary bg-primary/10 font-medium border-l-2 border-primary" 
                              : "text-muted-foreground hover:text-foreground hover:bg-muted"
                          )}
                        >
                          <span className="pl-7">{item.title}</span>
                        </Link>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </nav>
    </aside>
  )
}