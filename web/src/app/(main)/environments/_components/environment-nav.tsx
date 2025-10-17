'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { FileCode2, Settings, Key } from 'lucide-react'

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
    <div className="bg-background border-b">
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
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
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
