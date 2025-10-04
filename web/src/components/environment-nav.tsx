'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { FileCode2, Package, Settings, Key } from 'lucide-react'

interface TabItem {
  id: string
  label: string
  icon: React.ReactNode
  href: string
}

interface EnvironmentNavProps {
  environmentId: string
}

export function EnvironmentNav({ environmentId }: EnvironmentNavProps) {
  const pathname = usePathname()

  const tabs: TabItem[] = [
    {
      id: 'resources',
      label: 'Resources',
      icon: <Package className="h-4 w-4" />,
      href: `/environments/${environmentId}/resources`
    },
    {
      id: 'services',
      label: 'Services',
      icon: <FileCode2 className="h-4 w-4" />,
      href: `/environments/${environmentId}/services`
    },
    {
      id: 'configs',
      label: 'Config Management',
      icon: <Key className="h-4 w-4" />,
      href: `/environments/${environmentId}/configs`
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: <Settings className="h-4 w-4" />,
      href: `/environments/${environmentId}/settings`
    },
  ]

  return (
    <div className="bg-white border-b border-gray-200">
      <div className="mx-auto max-w-7xl px-6">
        <nav className="-mb-px flex space-x-8" aria-label="Tabs">
          {tabs.map((tab) => {
            const isActive = pathname.startsWith(tab.href)
            return (
              <Link
                key={tab.id}
                href={tab.href}
                className={`
                  flex items-center gap-2 px-1 py-4 text-sm font-medium border-b-2 transition-colors
                  ${isActive
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                  }
                `}
              >
                {tab.icon}
                {tab.label}
              </Link>
            )
          })}
        </nav>
      </div>
    </div>
  )
}
