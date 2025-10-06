'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Trash2 } from 'lucide-react'
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
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { deleteComposition } from '@/app/actions/composition.actions'
import type { Composition } from '@/types/composition'
import { EditCompositionSheet } from './edit-composition-sheet'
import { toast } from 'sonner'

interface CompositionRowActionsProps {
  composition: Composition
  namespace: string
  user: string
  isDeleting?: boolean
}

export function CompositionRowActions({ composition, namespace, user, isDeleting }: CompositionRowActionsProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  const handleDelete = async () => {
    startTransition(async () => {
      const result = await deleteComposition(namespace, composition.metadata.name, user)
      if (result.success) {
        toast.success('Composition deleted successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to delete composition')
      }
    })
  }

  if (isDeleting) {
    return <span className="text-xs text-gray-500">Deleting...</span>
  }

  return (
    <div className="flex items-center gap-2">
      <EditCompositionSheet composition={composition} namespace={namespace} user={user} />
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button variant="ghost" size="sm" className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50" disabled={isPending}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Composition</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete <strong>{composition.spec.displayName}</strong>? This action cannot be undone and will remove all associated resources.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-red-600 hover:bg-red-700">
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
