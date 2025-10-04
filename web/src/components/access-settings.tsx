'use client'

import { Button } from '@/components/ui/button'

export function AccessSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">Access Control</h3>
        <p className="text-sm text-gray-500">Manage environment visibility and access permissions</p>
      </div>
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Environment Visibility
            </label>
            <div className="space-y-2">
              <label className="flex items-center gap-2">
                <input type="radio" name="visibility" defaultChecked className="text-blue-600" />
                <span className="text-sm">Private - Only accessible to owner and invited members</span>
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
            <h4 className="text-sm font-medium text-gray-700 mb-2">Shared With</h4>
            <div className="space-y-2">
              <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
                <span className="text-sm">alice@team.com</span>
                <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">
                  Remove
                </Button>
              </div>
              <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
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
  )
}
