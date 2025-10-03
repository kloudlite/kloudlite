'use client'

import { Button } from '@/components/ui/button'

export function NetworkSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">Network Settings</h3>
        <p className="text-sm text-gray-500">Configure network and domain settings</p>
      </div>
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Custom Domain
            </label>
            <input
              type="text"
              className="w-full max-w-md px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
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
  )
}
