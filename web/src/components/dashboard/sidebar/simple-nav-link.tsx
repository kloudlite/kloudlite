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
        "flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors rounded-md group",
        "hover:bg-dashboard-hover hover:text-primary",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-background",
        isActive 
          ? "bg-primary/12 text-primary relative before:absolute before:-left-6 before:top-0 before:h-full before:w-1 before:bg-primary before:rounded-r" 
          : "text-muted-foreground"
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