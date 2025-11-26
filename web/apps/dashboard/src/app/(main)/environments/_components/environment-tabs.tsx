'use client'

import { useState } from 'react'
import { FileCode2, Package, Settings, Key } from 'lucide-react'

interface TabItem {
  id: string
  label: string
  icon: React.ReactNode
}

const tabs: TabItem[] = [
  { id: 'resources', label: 'Resources', icon: <Package className="h-4 w-4" /> },
  { id: 'services', label: 'Services', icon: <FileCode2 className="h-4 w-4" /> },
  { id: 'configs', label: 'Configs & Secrets', icon: <Key className="h-4 w-4" /> },
  { id: 'settings', label: 'Settings', icon: <Settings className="h-4 w-4" /> },
]

interface EnvironmentTabsProps {
  children: {
    resources?: React.ReactNode
    services?: React.ReactNode
    configs?: React.ReactNode
    settings?: React.ReactNode
  }
}

export function EnvironmentTabs({ children }: EnvironmentTabsProps) {
  const [activeTab, setActiveTab] = useState('resources')

  return (
    <div className="flex flex-1 flex-col">
      {/* Tab Navigation */}
      <div className="bg-card border-b">
        <div className="mx-auto max-w-7xl px-6">
          <nav className="-mb-px flex space-x-8" aria-label="Tabs">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 border-b-2 px-1 py-4 text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'border-info text-info'
                    : 'text-muted-foreground hover:border-border hover:text-foreground border-transparent'
                } `}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </nav>
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1">
        {activeTab === 'resources' && children.resources}
        {activeTab === 'services' && children.services}
        {activeTab === 'configs' && children.configs}
        {activeTab === 'settings' && children.settings}
      </div>
    </div>
  )
}
