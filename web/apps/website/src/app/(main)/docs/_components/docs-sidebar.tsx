'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { ThemeSwitcher } from '@/components/theme-switcher'
import type { Theme } from '@/lib/theme'
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from '@kloudlite/ui'
import { Menu } from 'lucide-react'
import { useState, useEffect } from 'react'

interface NavItem {
  title: string
  href: string
  items?: NavItem[]
}

interface DocsSidebarProps {
  initialTheme?: Theme
}

const navigation: NavItem[] = [
  {
    title: 'Introduction',
    href: '/docs',
    items: [
      { title: 'What is Kloudlite?', href: '/docs/introduction/what-is-kloudlite' },
      { title: 'Installation', href: '/docs/introduction/installation' },
      { title: 'Getting Started', href: '/docs/introduction/getting-started' },
      { title: 'Architecture', href: '/docs/architecture' },
    ],
  },
  {
    title: 'Workmachines',
    href: '/docs/workmachines',
    items: [
      { title: 'Overview', href: '/docs/workmachines' },
      { title: 'Access', href: '/docs/workmachines/access' },
    ],
  },
  {
    title: 'Environments',
    href: '/docs/concepts/environments',
    items: [
      { title: 'Overview', href: '/docs/concepts/environments' },
      { title: 'Services', href: '/docs/environment-internals/services' },
      { title: 'Configs & Secrets', href: '/docs/environment-internals/configs-secrets' },
    ],
  },
  {
    title: 'Workspaces',
    href: '/docs/concepts/workspaces',
    items: [
      { title: 'Overview', href: '/docs/concepts/workspaces' },
      { title: 'CLI Reference', href: '/docs/workspace-internals/cli' },
      { title: 'Env Connection', href: '/docs/workspace-internals/environment-connection' },
      { title: 'Service Intercepts', href: '/docs/workspace-internals/intercepts' },
      { title: 'Packages', href: '/docs/workspace-internals/packages' },
    ],
  },
  {
    title: 'References',
    href: 'https://github.com/kloudlite/kloudlite',
    items: [
      { title: 'GitHub', href: 'https://github.com/kloudlite/kloudlite' },
      { title: 'Issues', href: 'https://github.com/kloudlite/kloudlite/issues' },
    ],
  },
]

// Navigation content component (used by both mobile and desktop)
function NavigationContent({ pathname, onLinkClick }: { pathname: string; onLinkClick?: () => void }) {
  return (
    <div className="space-y-8">
      {navigation.map((section) => (
        <div key={section.href}>
          <Link
            href={section.href}
            onClick={onLinkClick}
            className={cn(
              'block px-3 py-2 text-sm font-semibold transition-all duration-100 active:translate-y-0.5',
              pathname === section.href
                ? 'text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {section.title}
          </Link>
          {section.items && (
            <ul className="mt-2 space-y-2 border-l pl-4">
              {section.items.map((item) => (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    onClick={onLinkClick}
                    className={cn(
                      'block py-1 pl-3 text-sm transition-all duration-100 active:translate-y-0.5',
                      pathname === item.href
                        ? 'text-foreground font-medium'
                        : 'text-muted-foreground hover:text-foreground',
                    )}
                  >
                    {item.title}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      ))}
    </div>
  )
}

export function DocsSidebar({ initialTheme = 'light' }: DocsSidebarProps) {
  const pathname = usePathname()
  const [open, setOpen] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <>
      {/* Mobile Menu Button - only render after mount to avoid hydration mismatch */}
      {mounted && (
        <div className="lg:hidden fixed bottom-6 right-6 z-50">
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
              <button className="bg-primary text-primary-foreground hover:bg-primary/90 flex h-14 w-14 items-center justify-center rounded-none shadow-lg transition-all duration-100 active:translate-y-0.5">
                <Menu className="h-6 w-6" />
              </button>
            </SheetTrigger>
            <SheetContent side="left" className="w-80 p-0">
              <div className="flex h-full flex-col">
                <div className="border-b px-4 py-3 sm:px-6 sm:py-4">
                  <SheetTitle className="text-lg font-semibold">Documentation</SheetTitle>
                </div>
                <nav className="flex-1 overflow-y-auto px-3 py-4 sm:px-4 sm:py-6">
                  <NavigationContent pathname={pathname} onLinkClick={() => setOpen(false)} />
                </nav>
                <div className="border-t px-4 py-3 sm:px-6 sm:py-4">
                  <div className="flex items-center justify-between">
                    <p className="text-muted-foreground text-xs">© 2024 Kloudlite</p>
                    <ThemeSwitcher initialTheme={initialTheme} />
                  </div>
                </div>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      )}

      {/* Desktop Sidebar */}
      <aside className="sticky top-16 hidden h-[calc(100vh-4rem)] w-64 flex-shrink-0 flex-col overflow-y-auto border-r bg-background lg:flex">
        <nav className="flex-1 px-3 py-6 sm:px-4 sm:py-8">
          <NavigationContent pathname={pathname} />
        </nav>

        {/* Sidebar Footer */}
        <div className="bg-background px-4 py-3 sm:px-6 sm:py-4">
          <div className="border-t mb-4"></div>
          <div className="flex items-center justify-between">
            <p className="text-muted-foreground text-xs">© 2024 Kloudlite</p>
            <ThemeSwitcher initialTheme={initialTheme} />
          </div>
        </div>
      </aside>
    </>
  )
}
