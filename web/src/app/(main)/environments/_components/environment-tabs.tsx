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
      <div className="border-b border-gray-200 bg-white">
        <div className="mx-auto max-w-7xl px-6">
          <nav className="-mb-px flex space-x-8" aria-label="Tabs">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 border-b-2 px-1 py-4 text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
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
