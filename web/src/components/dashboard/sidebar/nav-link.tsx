'use client'

import { Badge } from '@/components/ui/badge'
import { SidebarItem, SidebarLink } from '@/components/ui/sidebar'
import { cn } from '@/lib/utils'
import { usePathname } from 'next/navigation'
import type { NavItem } from '@/config/dashboard-navigation'

interface NavLinkProps {
  item: NavItem
  teamSlug: string
}

export function NavLink({ item, teamSlug }: NavLinkProps) {
  const pathname = usePathname()
  const href = `/${teamSlug}${item.href}`
  const isActive = pathname === href || (item.href && pathname.startsWith(href + '/'))
  const Icon = item.icon

  return (
    <SidebarItem>
      <SidebarLink
        href={href}
        active={isActive}
        icon={Icon}
      >
        <span className="flex-1">{item.label}</span>
        {item.badge && (
          <Badge 
            variant={item.badgeVariant || "secondary"} 
            className="text-xs"
          >
            {item.badge}
          </Badge>
        )}
      </SidebarLink>
    </SidebarItem>
  )
}