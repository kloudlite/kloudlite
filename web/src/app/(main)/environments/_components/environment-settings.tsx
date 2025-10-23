'use client'

import { useState } from 'react'
import { Settings, Users, Globe, Lock, Trash2, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'

type SettingSection = 'general' | 'access' | 'network' | 'security' | 'danger'

export function EnvironmentSettings() {
  const [activeSection, setActiveSection] = useState<SettingSection>('general')

  const sections = [
    { id: 'general' as const, label: 'General', icon: Settings },
    { id: 'access' as const, label: 'Access Control', icon: Users },
    { id: 'network' as const, label: 'Network', icon: Globe },
    { id: 'security' as const, label: 'Security', icon: Lock },
    { id: 'danger' as const, label: 'Danger Zone', icon: AlertTriangle },
  ]

  return (
    <div className="flex gap-6">
      {/* Left Navigation */}
      <div className="w-48 flex-shrink-0">
        <nav className="space-y-1">
          {sections.map((section) => {
            const Icon = section.icon
            return (
              <button
                key={section.id}
                onClick={() => setActiveSection(section.id)}
                className={`flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                  activeSection === section.id
                    ? 'bg-gray-100 font-medium text-gray-900'
                    : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                } ${section.id === 'danger' ? 'text-red-600 hover:bg-red-50' : ''} `}
              >
                <Icon className="h-4 w-4 flex-shrink-0" />
                {section.label}
              </button>
            )
          })}
        </nav>
      </div>

      {/* Content Area */}
      <div className="flex-1">
        {activeSection === 'general' && (
          <div className="space-y-4">
            <div className="mb-4">
              <h3 className="text-lg font-medium">General Settings</h3>
              <p className="text-sm text-gray-500">Configure basic environment properties</p>
            </div>
            <div className="rounded-lg border border-gray-200 bg-white p-6">
              <div className="space-y-4">
                <div>
                  <label className="mb-1 block text-sm font-medium text-gray-700">
                    Environment Name
                  </label>
                  <input
                    type="text"
                    className="w-full max-w-md rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                    defaultValue="my-dev-env"
                  />
                </div>

                <div>
                  <label className="mb-1 block text-sm font-medium text-gray-700">
                    Description
                  </label>
                  <textarea
                    className="w-full max-w-md rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                    rows={3}
                    defaultValue="Development environment for the main application"
                  />
                </div>

                <div className="pt-4">
                  <Button>Save Changes</Button>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeSection === 'access' && (
          <div className="space-y-4">
            <div className="mb-4">
              <h3 className="text-lg font-medium">Access Control</h3>
              <p className="text-sm text-gray-500">
                Manage environment visibility and access permissions
              </p>
            </div>
            <div className="rounded-lg border border-gray-200 bg-white p-6">
              <div className="space-y-4">
                <div>
                  <label className="mb-2 block text-sm font-medium text-gray-700">
                    Environment Visibility
                  </label>
                  <div className="space-y-2">
                    <label className="flex items-center gap-2">
                      <input
                        type="radio"
                        name="visibility"
                        defaultChecked
                        className="text-blue-600"
                      />
                      <span className="text-sm">
                        Private - Only accessible to owner and invited members
                      </span>
                    </label>
                    <label className="flex items-center gap-2">
                      <input type="radio" name="visibility" className="text-blue-600" />
                      <span className="text-sm">Team - Accessible to all team members</span>
                    </label>
                    <label className="flex items-center gap-2">
                      <input type="radio" name="visibility" className="text-blue-600" />
                      <span className="text-sm">Public - Accessible to everyone</span>
                    </label>
                  </div>
                </div>

                <div className="pt-2">
                  <h4 className="mb-2 text-sm font-medium text-gray-700">Shared With</h4>
                  <div className="space-y-2">
                    <div className="flex items-center justify-between rounded bg-gray-50 p-2">
                      <span className="text-sm">alice@team.com</span>
                      <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">
                        Remove
                      </Button>
                    </div>
                    <div className="flex items-center justify-between rounded bg-gray-50 p-2">
                      <span className="text-sm">bob@team.com</span>
                      <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">
                        Remove
                      </Button>
                    </div>
                  </div>
                  <Button variant="outline" size="sm" className="mt-3">
                    Add Member
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeSection === 'network' && (
          <div className="space-y-4">
            <div className="mb-4">
              <h3 className="text-lg font-medium">Network Settings</h3>
              <p className="text-sm text-gray-500">Configure network and domain settings</p>
            </div>
            <div className="rounded-lg border border-gray-200 bg-white p-6">
              <div className="space-y-4">
                <div>
                  <label className="mb-1 block text-sm font-medium text-gray-700">
                    Custom Domain
                  </label>
                  <input
                    type="text"
                    className="w-full max-w-md rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                    placeholder="app.example.com"
                  />
                </div>

                <div>
                  <label className="flex items-center gap-2">
                    <input type="checkbox" className="text-blue-600" />
                    <span className="text-sm">Enable HTTPS redirect</span>
                  </label>
                </div>

                <div>
                  <label className="flex items-center gap-2">
                    <input type="checkbox" className="text-blue-600" defaultChecked />
                    <span className="text-sm">Auto-generate SSL certificate</span>
                  </label>
                </div>

                <div className="pt-4">
                  <Button>Update Network Settings</Button>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeSection === 'security' && (
          <div className="space-y-4">
            <div className="mb-4">
              <h3 className="text-lg font-medium">Security</h3>
              <p className="text-sm text-gray-500">Security policies and encryption settings</p>
            </div>
            <div className="rounded-lg border border-gray-200 bg-white p-6">
              <div className="space-y-4">
                <div>
                  <label className="flex items-center gap-2">
                    <input type="checkbox" className="text-blue-600" defaultChecked />
                    <span className="text-sm">Enable network policies</span>
                  </label>
                </div>

                <div>
                  <label className="flex items-center gap-2">
                    <input type="checkbox" className="text-blue-600" defaultChecked />
                    <span className="text-sm">Encrypt secrets at rest</span>
                  </label>
                </div>

                <div>
                  <label className="flex items-center gap-2">
                    <input type="checkbox" className="text-blue-600" />
                    <span className="text-sm">Enable pod security policies</span>
                  </label>
                </div>

                <div className="pt-4">
                  <Button>Update Security Settings</Button>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeSection === 'danger' && (
          <div className="space-y-4">
            <div className="mb-4">
              <h3 className="text-lg font-medium text-red-900">Danger Zone</h3>
              <p className="text-sm text-red-600">Irreversible and destructive actions</p>
            </div>
            <div className="rounded-lg border border-red-200 bg-red-50 p-6">
              <div className="space-y-4">
                <div>
                  <p className="mb-3 text-sm text-red-700">
                    Once you delete an environment, there is no going back. All resources will be
                    permanently removed.
                  </p>
                  <Button variant="outline" className="border-red-500 text-red-600 hover:bg-red-50">
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete Environment
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
