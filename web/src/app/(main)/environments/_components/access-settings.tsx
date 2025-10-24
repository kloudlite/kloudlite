'use client'

import { Button } from '@/components/ui/button'

export function AccessSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">Access Control</h3>
        <p className="text-muted-foreground text-sm">
          Manage environment visibility and access permissions
        </p>
      </div>
      <div className="bg-card rounded-lg border p-6">
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium">Environment Visibility</label>
            <div className="space-y-2">
              <label className="flex items-center gap-2">
                <input type="radio" name="visibility" defaultChecked className="text-info" />
                <span className="text-sm">
                  Private - Only accessible to owner and invited members
                </span>
              </label>
              <label className="flex items-center gap-2">
                <input type="radio" name="visibility" className="text-info" />
                <span className="text-sm">Team - Accessible to all team members</span>
              </label>
              <label className="flex items-center gap-2">
                <input type="radio" name="visibility" className="text-info" />
                <span className="text-sm">Public - Accessible to everyone</span>
              </label>
            </div>
          </div>

          <div className="pt-2">
            <h4 className="mb-2 text-sm font-medium">Shared With</h4>
            <div className="space-y-2">
              <div className="bg-muted/50 flex items-center justify-between rounded p-2">
                <span className="text-sm">alice@team.com</span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-destructive hover:text-destructive/80"
                >
                  Remove
                </Button>
              </div>
              <div className="bg-muted/50 flex items-center justify-between rounded p-2">
                <span className="text-sm">bob@team.com</span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-destructive hover:text-destructive/80"
                >
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
