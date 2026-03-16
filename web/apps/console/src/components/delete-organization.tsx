'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, Trash2 } from 'lucide-react'
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
  Button,
  Input,
  Label,
} from '@kloudlite/ui'

interface DeleteOrganizationProps {
  orgId: string
  orgName: string
}

export function DeleteOrganization({ orgId, orgName }: DeleteOrganizationProps) {
  const router = useRouter()
  const [confirmText, setConfirmText] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')

  const confirmed = confirmText === orgName

  const handleDelete = async () => {
    if (!confirmed) return
    setDeleting(true)
    setError('')
    try {
      const res = await fetch(`/api/orgs/${orgId}`, { method: 'DELETE' })
      if (!res.ok) {
        const data = await res.json()
        setError(data.error || 'Failed to delete organization')
        return
      }
      // Redirect to installations — getSelectedOrg will fall back to another org or auto-create
      router.push('/installations')
      router.refresh()
    } catch {
      setError('Something went wrong')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="border border-destructive/20 rounded-lg p-6 bg-destructive/[0.03]">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-semibold text-destructive">Delete Organization</h3>
          <p className="text-muted-foreground text-sm mt-1 max-w-lg">
            Permanently delete this organization, all its installations, team members, invitations,
            and billing data. Active subscriptions will be cancelled. This action cannot be undone.
          </p>
        </div>
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" size="sm">
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Organization
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete &ldquo;{orgName}&rdquo;?</AlertDialogTitle>
              <AlertDialogDescription>
                This will permanently delete the organization and all associated data including
                installations, team memberships, and billing. Active Stripe subscriptions will be
                cancelled. This action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <div className="py-4">
              <Label htmlFor="confirm-delete" className="text-sm">
                Type <span className="font-semibold">{orgName}</span> to confirm
              </Label>
              <Input
                id="confirm-delete"
                className="mt-2"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder={orgName}
              />
              {error && <p className="text-sm text-destructive mt-2">{error}</p>}
            </div>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => { setConfirmText(''); setError('') }}>
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleDelete}
                disabled={!confirmed || deleting}
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              >
                {deleting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                Delete Organization
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  )
}
