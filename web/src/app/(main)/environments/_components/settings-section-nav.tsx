'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Settings, Users, Globe, Lock, AlertTriangle } from 'lucide-react'

interface SettingsSectionNavProps {
  environmentId: string
}

export function SettingsSectionNav({ environmentId }: SettingsSectionNavProps) {
  const pathname = usePathname()

  const sections = [
    {
      id: 'general',
      label: 'General',
      icon: Settings,
      href: `/environments/${environmentId}/settings/general`,
    },
    {
      id: 'access',
      label: 'Access Control',
      icon: Users,
      href: `/environments/${environmentId}/settings/access`,
    },
    {
      id: 'network',
      label: 'Network',
      icon: Globe,
      href: `/environments/${environmentId}/settings/network`,
    },
    {
      id: 'security',
      label: 'Security',
      icon: Lock,
      href: `/environments/${environmentId}/settings/security`,
    },
    {
      id: 'danger',
      label: 'Danger Zone',
      icon: AlertTriangle,
      href: `/environments/${environmentId}/settings/danger`,
    },
  ]

  return (
    <div className="w-48 flex-shrink-0">
      <nav className="space-y-1">
        {sections.map((section) => {
          const Icon = section.icon
          const isActive = pathname === section.href
          const isDanger = section.id === 'danger'
          return (
            <Link
              key={section.id}
              href={section.href}
              className={`w-full flex items-center gap-3 px-3 py-2 text-sm rounded-md transition-colors ${
                isActive
                  ? isDanger
                    ? 'bg-red-50 text-red-700 font-medium'
                    : 'bg-gray-100 text-gray-900 font-medium'
                  : isDanger
                  ? 'text-red-600 hover:bg-red-50'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
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
