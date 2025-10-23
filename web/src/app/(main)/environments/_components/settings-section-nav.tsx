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
              className={`flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                isActive
                  ? isDanger
                    ? 'bg-red-50 font-medium text-red-700 dark:bg-red-900/20 dark:text-red-400'
                    : 'bg-accent text-accent-foreground font-medium'
                  : isDanger
                    ? 'text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20'
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
