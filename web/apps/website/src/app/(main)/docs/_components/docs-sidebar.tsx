'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { ThemeSwitcher } from '@kloudlite/ui'
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from '@kloudlite/ui'
import { Menu } from 'lucide-react'
import { useState, useEffect } from 'react'

interface NavItem {
  title: string
  href: string
  items?: NavItem[]
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
          <h3 className="px-4 mb-3 text-[11px] font-bold uppercase tracking-widest text-muted-foreground/80">
            {section.title}
          </h3>
          {section.items && (
            <ul className="space-y-0.5">
              {section.items.map((item) => (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    onClick={onLinkClick}
                    className={cn(
                      'group relative block px-4 py-2.5 text-[13px] font-medium transition-[color,background-color] duration-300 rounded-sm overflow-hidden',
                      pathname === item.href
                        ? 'text-primary bg-primary/[0.08]'
                        : 'text-muted-foreground hover:text-foreground hover:bg-foreground/[0.04]',
                    )}
                  >
                    {/* Active indicator bar */}
                    {pathname === item.href && (
                      <span className="absolute left-0 top-0 bottom-0 w-[3px] bg-primary" />
                    )}

                    {/* Hover indicator bar (only for non-active items) */}
                    {pathname !== item.href && (
                      <span className="absolute left-0 top-0 bottom-0 w-[3px] bg-foreground/20 scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-center" />
                    )}

                    <span className="relative z-10">{item.title}</span>
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

export function DocsSidebar() {
  const pathname = usePathname()
  const [open, setOpen] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      setMounted(true)
    })
    return () => cancelAnimationFrame(frame)
  }, [])

  return (
    <>
      {/* Mobile Menu Button - only render after mount to avoid hydration mismatch */}
      {mounted && (
        <div className="lg:hidden fixed bottom-4 right-4 sm:bottom-6 sm:right-6 z-50">
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
              <button className="bg-primary text-primary-foreground hover:bg-primary/90 flex h-14 w-14 items-center justify-center shadow-lg transition-colors">
                <Menu className="h-6 w-6" />
              </button>
            </SheetTrigger>
            <SheetContent side="left" className="w-[calc(100vw-2rem)] max-w-80 p-0 border-r border-foreground/10">
              <div className="flex h-full flex-col bg-background">
                <div className="border-b border-foreground/10 px-6 py-5">
                  <SheetTitle className="text-base font-bold">Documentation</SheetTitle>
                </div>
                <nav className="flex-1 overflow-y-auto px-3 py-6">
                  <NavigationContent pathname={pathname} onLinkClick={() => setOpen(false)} />
                </nav>
                <div className="border-t border-foreground/10 px-6 py-4 bg-foreground/[0.01]">
                  <div className="flex items-center justify-between">
                    <p className="text-muted-foreground text-xs font-medium">© 2026 Kloudlite</p>
                    <ThemeSwitcher />
                  </div>
                </div>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      )}

      {/* Desktop Sidebar */}
      <aside className="sticky top-0 hidden w-64 flex-shrink-0 border-r border-foreground/10 bg-background lg:block h-[calc(100vh-4rem)]">
        <div className="flex flex-col h-full">
          <nav className="flex-1 overflow-y-auto px-3 py-8">
            <NavigationContent pathname={pathname} />
          </nav>

          {/* Sidebar Footer */}
          <div className="border-t border-foreground/10 px-6 py-4 bg-foreground/[0.01]">
            <div className="flex items-center justify-between">
              <p className="text-muted-foreground text-xs font-medium">© 2026 Kloudlite</p>
              <ThemeSwitcher />
            </div>
          </div>
        </div>
      </aside>
    </>
  )
}
