'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { User, CreditCard } from 'lucide-react'
import { cn } from '@/lib/utils'

export function InstallationSettingsTabs() {
  const pathname = usePathname()

  const tabs = [
    {
      id: 'profile',
      label: 'Profile',
      icon: User,
      href: '/installations/settings/profile',
    },
    {
      id: 'billing',
      label: 'Billing',
      icon: CreditCard,
      href: '/installations/settings/billing',
    },
  ]

  return (
    <div className="border-b">
      <nav className="-mb-px flex gap-6">
        {tabs.map((tab) => {
          const Icon = tab.icon
          const isActive = pathname === tab.href
          return (
            <Link
              key={tab.id}
              href={tab.href}
              className={cn(
                'flex items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium transition-colors',
                isActive
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:border-muted-foreground/30 hover:text-foreground',
              )}
            >
              <Icon className="h-4 w-4" />
              {tab.label}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
