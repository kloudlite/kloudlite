'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Package2, FileCode } from 'lucide-react'

interface ResourceSectionNavProps {
  environmentId: string
}

export function ResourceSectionNav({ environmentId }: ResourceSectionNavProps) {
  const pathname = usePathname()

  const sections = [
    {
      id: 'helmcharts',
      label: 'Helm Charts',
      icon: Package2,
      href: `/environments/${environmentId}/resources/helmcharts`,
    },
    {
      id: 'compositions',
      label: 'Compositions',
      icon: FileCode,
      href: `/environments/${environmentId}/resources/compositions`,
    },
  ]

  return (
    <div className="w-48 flex-shrink-0">
      <nav className="space-y-1">
        {sections.map((section) => {
          const Icon = section.icon
          const isActive = pathname === section.href
          return (
            <Link
              key={section.id}
              href={section.href}
              className={`w-full flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                isActive
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
              }`}
            >
              <Icon className="h-4 w-4" />
              {section.label}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
