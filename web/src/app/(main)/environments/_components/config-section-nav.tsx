'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { FileText, File } from 'lucide-react'

interface ConfigSectionNavProps {
  environmentId: string
}

export function ConfigSectionNav({ environmentId }: ConfigSectionNavProps) {
  const pathname = usePathname()

  const sections = [
    {
      id: 'envvars',
      label: 'Envvars',
      icon: FileText,
      href: `/environments/${environmentId}/configs/envvars`,
    },
    {
      id: 'config-files',
      label: 'Config Files',
      icon: File,
      href: `/environments/${environmentId}/configs/config-files`,
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
              className={`w-full flex items-center gap-3 px-3 py-2 text-sm rounded-md transition-colors ${
                isActive
                  ? 'bg-accent text-accent-foreground font-medium'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
              }`}
            >
              <Icon className="h-4 w-4 flex-shrink-0" />
              {section.label}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
