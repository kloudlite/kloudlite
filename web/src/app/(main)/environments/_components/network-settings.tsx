'use client'

import { Button } from '@/components/ui/button'

export function NetworkSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">Network Settings</h3>
        <p className="text-muted-foreground text-sm">Configure network and domain settings</p>
      </div>
      <div className="bg-card rounded-lg border p-6">
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium">Custom Domain</label>
            <input
              type="text"
              className="w-full max-w-md rounded-md border px-3 py-2 focus:ring-2 focus:ring-info focus:outline-none"
              placeholder="app.example.com"
            />
          </div>

          <div>
            <label className="flex items-center gap-2">
              <input type="checkbox" className="text-info" />
              <span className="text-sm">Enable HTTPS redirect</span>
            </label>
          </div>

          <div>
            <label className="flex items-center gap-2">
              <input type="checkbox" className="text-info" defaultChecked />
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
