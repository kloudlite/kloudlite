'use client'

import { Button } from '@/components/ui/button'

export function SecuritySettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">Security</h3>
        <p className="text-muted-foreground text-sm">Security policies and encryption settings</p>
      </div>
      <div className="bg-card rounded-lg border p-6">
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
  )
}
