'use client'

import { Button } from '@/components/ui/button'

export function GeneralSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">General Settings</h3>
        <p className="text-muted-foreground text-sm">Configure basic environment properties</p>
      </div>
      <div className="bg-card rounded-lg border p-6">
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium">Environment Name</label>
            <input
              type="text"
              className="w-full max-w-md rounded-md border px-3 py-2 focus:ring-2 focus:ring-info focus:outline-none"
              defaultValue="my-dev-env"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium">Description</label>
            <textarea
              className="w-full max-w-md rounded-md border px-3 py-2 focus:ring-2 focus:ring-info focus:outline-none"
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
