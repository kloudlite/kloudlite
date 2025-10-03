'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { FileText, Lock, File } from 'lucide-react'

interface ConfigSectionNavProps {
  environmentId: string
}

export function ConfigSectionNav({ environmentId }: ConfigSectionNavProps) {
  const pathname = usePathname()

  const sections = [
    {
      id: 'configmaps',
      label: 'Config Maps',
      icon: FileText,
      href: `/environments/${environmentId}/configs/configmaps`,
      count: 7
    },
    {
      id: 'secrets',
      label: 'Secrets',
      icon: Lock,
      href: `/environments/${environmentId}/configs/secrets`,
      count: 5
    },
    {
      id: 'files',
      label: 'File Configs',
      icon: File,
      href: `/environments/${environmentId}/configs/files`,
      count: 4
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
                  ? 'bg-gray-100 text-gray-900 font-medium'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              }`}
            >
              <Icon className="h-4 w-4 flex-shrink-0" />
              {section.label}
              <span className="ml-auto text-xs text-gray-500">{section.count}</span>
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
