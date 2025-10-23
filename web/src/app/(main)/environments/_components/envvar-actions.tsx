'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Pencil, Trash2, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { deleteEnvVar } from '@/app/actions/environment-config'
import type { EnvVar } from '@/types/environment'
import { EditEnvVarSheet } from './edit-envvar-sheet'
import { toast } from 'sonner'

interface EnvVarActionsProps {
  envVar: EnvVar
  environmentId: string
}

export function EnvVarActions({ envVar, environmentId }: EnvVarActionsProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showEditSheet, setShowEditSheet] = useState(false)

  const handleDelete = () => {
    startTransition(async () => {
      try {
        await deleteEnvVar(environmentId, envVar.key)
        toast.success('Environment variable deleted successfully')
        setShowDeleteDialog(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete environment variable')
      }
    })
  }

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        className="text-gray-600 hover:text-gray-700"
        onClick={() => setShowEditSheet(true)}
      >
        <Pencil className="h-4 w-4" />
      </Button>
      <Button
        variant="ghost"
        size="sm"
        className="text-red-600 hover:text-red-700"
        onClick={() => setShowDeleteDialog(true)}
      >
        <Trash2 className="h-4 w-4" />
      </Button>

      {/* Edit Sheet */}
      <EditEnvVarSheet
        environmentId={environmentId}
        envVar={envVar}
        open={showEditSheet}
        onOpenChange={setShowEditSheet}
        onSuccess={() => router.refresh()}
      />

      {/* Delete Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Envvar</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the envvar{' '}
              <span className="font-mono font-semibold">{envVar.key}</span>? This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isPending}
              className="bg-red-600 hover:bg-red-700"
            >
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
