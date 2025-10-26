'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { ThemeSwitcher } from '@/components/theme-switcher'
import type { Theme } from '@/lib/theme'

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
      { title: 'Getting Started', href: '/docs' },
      { title: 'Architecture', href: '/docs/architecture' },
    ],
  },
  {
    title: 'Core Concepts',
    href: '/docs/concepts',
    items: [
      { title: 'Workspaces', href: '/docs/concepts/workspaces' },
      { title: 'Environments', href: '/docs/concepts/environments' },
    ],
  },
  {
    title: 'Workspace Internals',
    href: '/docs/workspace-internals',
    items: [
      { title: 'Environment Connection', href: '/docs/workspace-internals/environment-connection' },
      { title: 'Service Intercepts', href: '/docs/workspace-internals/intercepts' },
      { title: 'Package Management', href: '/docs/workspace-internals/packages' },
      { title: 'CLI Reference', href: '/docs/workspace-internals/cli' },
    ],
  },
  {
    title: 'Installation',
    href: '/docs/installation',
    items: [
      { title: 'BYOC Setup', href: '/docs/installation/byoc' },
      { title: 'Cloud vs BYOC', href: '/docs/installation/cloud-vs-byoc' },
    ],
  },
  {
    title: 'Help & Support',
    href: '/docs/help',
    items: [
      { title: 'FAQ', href: '/docs/faq' },
      { title: 'Contact Us', href: '/contact' },
      { title: 'Report Issue', href: 'https://github.com/kloudlite/kloudlite/issues/new' },
      {
        title: 'Feature Request',
        href: 'https://github.com/kloudlite/kloudlite/issues/new?labels=enhancement',
      },
    ],
  },
]

export function DocsSidebar({ initialTheme = 'light' }: DocsSidebarProps) {
  const pathname = usePathname()

  return (
    <aside className="sticky top-16 hidden h-[calc(100vh-4rem)] w-64 flex-shrink-0 flex-col overflow-y-auto border-r lg:flex">
      <nav className="flex-1 px-4 py-8">
        <div className="space-y-8">
          {navigation.map((section) => (
            <div key={section.href}>
              <Link
                href={section.href}
                className={cn(
                  'block px-3 py-2 text-sm font-semibold transition-colors',
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
                        className={cn(
                          'block py-1 pl-3 text-sm transition-colors',
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
      </nav>

      {/* Sidebar Footer */}
      <div className="bg-muted border-t px-4 py-3">
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground text-xs">© 2024 Kloudlite</p>
          <ThemeSwitcher initialTheme={initialTheme} />
        </div>
      </div>
    </aside>
  )
}
