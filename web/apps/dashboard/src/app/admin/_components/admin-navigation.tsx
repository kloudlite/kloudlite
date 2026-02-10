'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Users, Settings, Shield } from 'lucide-react'

interface AdminNavigationProps {
  isKloudliteCloud?: boolean
}

export function AdminNavigation({ isKloudliteCloud }: AdminNavigationProps) {
  const pathname = usePathname()

  const allNavItems = [
    { href: '/admin/users', label: 'User Management', icon: Users },
    { href: '/admin/machine-configs', label: 'Machine Configs', icon: Settings },
    { href: '/admin/oauth-providers', label: 'OAuth Providers', icon: Shield },
  ]

  const navItems = isKloudliteCloud
    ? allNavItems.filter((item) => item.href !== '/admin/machine-configs')
    : allNavItems

  return (
    <nav className="hidden items-center gap-1 md:flex">
      {navItems.map((item) => {
        const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)
        const Icon = item.icon

        return (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors ${
              isActive
                ? 'bg-muted text-foreground font-medium'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            }`}
          >
            <Icon className="h-4 w-4" />
            {item.label}
          </Link>
        )
      })}
    </nav>
  )
}
