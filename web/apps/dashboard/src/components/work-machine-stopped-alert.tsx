'use client'

import Link from 'next/link'
import { Power } from 'lucide-react'
import { Button } from '@kloudlite/ui'

export function WorkMachineStoppedAlert() {
  return (
    <div className="mb-6 rounded-lg border border-amber-200 dark:border-amber-900/50 bg-amber-50 dark:bg-amber-950/30 p-4">
      <div className="flex items-start gap-4">
        <div className="flex-shrink-0 rounded-full bg-amber-100 dark:bg-amber-900/50 p-2">
          <Power className="h-4 w-4 text-amber-600 dark:text-amber-400" />
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-amber-900 dark:text-amber-100">
            WorkMachine is stopped
          </h3>
          <p className="text-sm text-amber-700 dark:text-amber-300 mt-1">
            Start your WorkMachine to access workspaces and environments.
          </p>
        </div>
        <Link href="/" className="flex-shrink-0">
          <Button size="sm" variant="outline" className="border-amber-300 dark:border-amber-700 hover:bg-amber-100 dark:hover:bg-amber-900/50">
            <Power className="mr-2 h-3.5 w-3.5" />
            Start
          </Button>
        </Link>
      </div>
    </div>
  )
}
