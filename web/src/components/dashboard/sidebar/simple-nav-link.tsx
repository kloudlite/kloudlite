'use client'

import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { usePathname } from 'next/navigation'
import type { NavItem } from '@/config/dashboard-navigation'

interface SimpleNavLinkProps {
  item: NavItem
  teamSlug: string
}

export function SimpleNavLink({ item, teamSlug }: SimpleNavLinkProps) {
  const pathname = usePathname()
  const href = `/${teamSlug}${item.href}`
  const isActive = pathname === href || (item.href && pathname.startsWith(href + '/'))
  const Icon = item.icon

  return (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors",
        "hover:bg-muted hover:text-foreground",
        isActive && "bg-primary/10 text-primary relative before:absolute before:-left-6 before:top-0 before:h-full before:w-1 before:bg-primary"
      )}
    >
      <Icon className="size-4 flex-shrink-0" />
      <span className="flex-1">{item.label}</span>
      {item.badge && (
        <Badge 
          variant={item.badgeVariant || "secondary"} 
          className="text-xs"
        >
          {item.badge}
        </Badge>
      )}
    </Link>
  )
}