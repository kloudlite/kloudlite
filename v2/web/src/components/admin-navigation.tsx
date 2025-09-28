'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ArrowLeft } from 'lucide-react'

export function AdminNavigation() {
  const pathname = usePathname()

  const navItems = [
    { href: '/admin/users', label: 'Users' },
    { href: '/admin/machine-configs', label: 'Machine Configurations' },
  ]

  return (
    <header className="border-b bg-white">
      <div className="mx-auto max-w-7xl px-6">
        <div className="flex h-16 items-center justify-between">
          {/* Left side - Logo and Navigation (like main app) */}
          <div className="flex items-center gap-8">
            <Link href="/" className="text-lg font-medium flex items-center gap-2">
              <ArrowLeft className="h-4 w-4" />
              Administration
            </Link>

            {/* Main Navigation */}
            <nav className="hidden md:flex items-center gap-1">
              {navItems.map((item) => {
                const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`px-3 py-2 text-sm rounded-md transition-colors ${
                      isActive
                        ? 'bg-gray-100 text-gray-900 font-medium'
                        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                    }`}
                  >
                    {item.label}
                  </Link>
                )
              })}
            </nav>
          </div>

          {/* Right side - empty to match main app layout */}
          <div />
        </div>
      </div>
    </header>
  )
}