'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Container } from 'lucide-react'

const navItems = [
  {
    href: '/artifacts/container-images',
    label: 'Container Images',
    icon: Container,
  },
  // Future artifact types can be added here:
  // { href: '/artifacts/helm-charts', label: 'Helm Charts', icon: Package, comingSoon: true },
  // { href: '/artifacts/packages', label: 'Packages', icon: Archive, comingSoon: true },
]

export function ArtifactsNav() {
  const pathname = usePathname()

  return (
    <nav className="border-b">
      <div className="flex gap-4">
        {navItems.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)
          const Icon = item.icon

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
