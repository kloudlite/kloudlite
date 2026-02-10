'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Trash2, Loader2, AlertTriangle } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { deleteEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'

interface DangerSettingsProps {
  environmentId: string
  environmentName: string
}

export function DangerSettings({ environmentId, environmentName }: DangerSettingsProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [confirmText, setConfirmText] = useState('')

  const handleDelete = () => {
    startTransition(async () => {
      try {
        const result = await deleteEnvironment(environmentId)
        if (result.success) {
          toast.success('Environment deleted successfully')
          router.push('/environments')
        } else {
          toast.error(result.error || 'Failed to delete environment')
        }
      } catch {
        toast.error('Failed to delete environment')
      } finally {
        setShowDeleteDialog(false)
      }
    })
  }

  const isConfirmValid = confirmText === environmentName

  return (
    <>
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
                onClick={() => setShowDeleteDialog(true)}
                className="border-red-500 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/30"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Environment
              </Button>
            </div>
          </div>
        </div>
      </div>

      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <AlertTriangle className="h-5 w-5" />
              Delete Environment
            </DialogTitle>
            <DialogDescription>
              This action cannot be undone. This will permanently delete the environment
              and all associated resources.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <p className="text-sm">
              Please type <span className="font-mono font-semibold">{environmentName}</span> to confirm.
            </p>
            <Input
              placeholder="Type environment name to confirm"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              disabled={isPending}
              className="font-mono"
            />
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteDialog(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={!isConfirmValid || isPending}
            >
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete Environment
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
