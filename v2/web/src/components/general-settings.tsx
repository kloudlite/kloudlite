'use client'

import { Button } from '@/components/ui/button'

export function GeneralSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">General Settings</h3>
        <p className="text-sm text-gray-500">Configure basic environment properties</p>
      </div>
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Environment Name
            </label>
            <input
              type="text"
              className="w-full max-w-md px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              defaultValue="my-dev-env"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <textarea
              className="w-full max-w-md px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
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
  )
}
