'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Users, Settings, Shield } from 'lucide-react'

export function AdminNavigation() {
  const pathname = usePathname()

  const navItems = [
    { href: '/admin/users', label: 'User Management', icon: Users },
    { href: '/admin/machine-configs', label: 'Machine Configs', icon: Settings },
    { href: '/admin/oauth-providers', label: 'OAuth Providers', icon: Shield },
  ]

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
                ? 'bg-gray-100 font-medium text-gray-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
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
