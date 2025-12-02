'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Container, Box } from 'lucide-react'

interface NavItem {
  href: string
  label: string
  icon: typeof Container
  comingSoon?: boolean
}

const navItems: NavItem[] = [
  {
    href: '/artifacts/container-repos',
    label: 'Container Repos',
    icon: Container,
  },
  {
    href: '/artifacts/model-repos',
    label: 'Model Repos',
    icon: Box,
    comingSoon: true,
  },
]

export function ArtifactsNav() {
  const pathname = usePathname()

  return (
    <nav className="border-b">
      <div className="flex gap-4">
        {navItems.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)
          const Icon = item.icon

          if (item.comingSoon) {
            return (
              <div
                key={item.href}
                className="flex items-center gap-2 border-b-2 border-transparent px-1 py-3 text-sm font-medium text-muted-foreground/60 cursor-not-allowed"
              >
                <Icon className="h-4 w-4" />
                {item.label}
                <span className="rounded-full bg-muted px-2 py-0.5 text-xs">Coming Soon</span>
              </div>
            )
          }

          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium transition-colors ${
                isActive
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:border-muted-foreground/50 hover:text-foreground'
              }`}
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
