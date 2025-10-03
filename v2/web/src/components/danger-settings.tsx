'use client'

import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function DangerSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium text-red-900">Danger Zone</h3>
        <p className="text-sm text-red-600">Irreversible and destructive actions</p>
      </div>
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <div className="space-y-4">
          <div>
            <p className="text-sm text-red-700 mb-3">
              Once you delete an environment, there is no going back. All resources will be permanently removed.
            </p>
            <Button variant="outline" className="border-red-500 text-red-600 hover:bg-red-50">
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Environment
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
