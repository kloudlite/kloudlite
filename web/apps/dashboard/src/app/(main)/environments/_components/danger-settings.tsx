'use client'

import { Trash2 } from 'lucide-react'
import { Button } from '@kloudlite/ui'

export function DangerSettings() {
  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium text-red-900 dark:text-red-400">Danger Zone</h3>
        <p className="text-sm text-red-600 dark:text-red-400">
          Irreversible and destructive actions
        </p>
      </div>
      <div className="rounded-lg border border-red-200 bg-red-50 p-6 dark:border-red-800 dark:bg-red-900/20">
        <div className="space-y-4">
          <div>
            <p className="mb-3 text-sm text-red-700 dark:text-red-300">
              Once you delete an environment, there is no going back. All resources will be
              permanently removed.
            </p>
            <Button
              variant="outline"
              className="border-red-500 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/30"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete Environment
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
