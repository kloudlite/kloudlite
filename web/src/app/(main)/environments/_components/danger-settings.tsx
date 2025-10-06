'use client'

import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function DangerSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium text-red-900 dark:text-red-400">Danger Zone</h3>
        <p className="text-sm text-red-600 dark:text-red-400">Irreversible and destructive actions</p>
      </div>
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6">
        <div className="space-y-4">
          <div>
            <p className="text-sm text-red-700 dark:text-red-300 mb-3">
              Once you delete an environment, there is no going back. All resources will be permanently removed.
            </p>
            <Button variant="outline" className="border-red-500 dark:border-red-700 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30">
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Environment
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
