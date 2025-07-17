'use client'

import { ChevronRight } from 'lucide-react'
import { Link } from '@/components/ui/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'

interface DashboardBreadcrumbProps {
  teamName: string
}

export function DashboardBreadcrumb({ teamName }: DashboardBreadcrumbProps) {
  const pathname = usePathname()
  const segments = pathname.split('/').filter(Boolean)
  
  // Remove team slug from segments
  const pathSegments = segments.slice(1)
  
  // Build breadcrumb items
  const breadcrumbs = [
    { label: teamName, href: `/${segments[0]}` },
    ...pathSegments.map((segment, index) => {
      const href = `/${segments[0]}/${pathSegments.slice(0, index + 1).join('/')}`
      const label = segment
        .split('-')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ')
      return { label, href }
    })
  ]

  if (breadcrumbs.length === 1) {
    return (
      <div className="flex items-center gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <span className="text-sm text-muted-foreground">— Welcome back</span>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      <nav className="flex items-center text-sm">
        {breadcrumbs.map((crumb, index) => (
          <div key={crumb.href} className="flex items-center">
            {index > 0 && (
              <ChevronRight className="h-4 w-4 mx-1 text-muted-foreground/50" />
            )}
            {index === breadcrumbs.length - 1 ? (
              <h1 className="text-2xl font-semibold tracking-tight">
                {crumb.label}
              </h1>
            ) : (
              <Link
                href={crumb.href}
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                {crumb.label}
              </Link>
            )}
          </div>
        ))}
      </nav>
    </div>
  )
}