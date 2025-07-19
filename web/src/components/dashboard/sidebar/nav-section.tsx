'use client'

import { SimpleNavLink } from './simple-nav-link'
import { mainNavItems } from '@/config/dashboard-navigation'
import { ScrollArea } from '@/components/ui/scroll-area'

interface NavSectionProps {
  teamSlug: string
}

export function NavSection({ teamSlug }: NavSectionProps) {
  return (
    <nav className="flex-1 min-h-0 overflow-hidden">
      <ScrollArea className="h-full" scrollbarVariant="minimal" fadeScrollbar>
        <div className="px-4 py-5 space-y-1">
          {mainNavItems.map((item) => (
            <SimpleNavLink key={item.href} item={item} teamSlug={teamSlug} />
          ))}
        </div>
      </ScrollArea>
    </nav>
  )
}