'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@kloudlite/lib'
import { CreditCard, Receipt } from 'lucide-react'

interface BillingSidebarProps {
  installationId: string
}

const links = [
  { label: 'Subscription', segment: 'subscription', icon: CreditCard },
  { label: 'Invoices', segment: 'invoices', icon: Receipt },
]

export function BillingSidebar({ installationId }: BillingSidebarProps) {
  const pathname = usePathname()
  const base = `/installations/${installationId}/billing`

  return (
    <nav className="flex flex-col gap-1">
      {links.map((link) => {
        const href = `${base}/${link.segment}`
        const isActive = pathname === href
        const Icon = link.icon
        return (
          <Link
            key={link.segment}
            href={href}
            className={cn(
              'flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              isActive
                ? 'bg-primary/10 text-primary'
                : 'text-muted-foreground hover:bg-foreground/5 hover:text-foreground',
            )}
          >
            <Icon className="h-4 w-4" />
            {link.label}
          </Link>
        )
      })}
    </nav>
  )
}
