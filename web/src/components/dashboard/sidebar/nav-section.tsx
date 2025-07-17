'use client'

import { SimpleNavLink } from './simple-nav-link'
import { mainNavItems } from '@/config/dashboard-navigation'

interface NavSectionProps {
  teamSlug: string
}

export function NavSection({ teamSlug }: NavSectionProps) {
  return (
    <nav className="flex-1 min-h-0 overflow-y-auto">
      <div className="px-4 py-5 space-y-1">
        {mainNavItems.map((item) => (
          <SimpleNavLink key={item.href} item={item} teamSlug={teamSlug} />
        ))}
      </div>
    </nav>
  )
}