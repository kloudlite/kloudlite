import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { getWorkspaceData } from '../workspace-data'
import { WorkspaceDescriptionForm } from '../../_components/workspace-description-form'
import { WorkspaceDangerZone } from '../../_components/workspace-danger-zone'
import { Info, User, Server } from 'lucide-react'
import type { PageProps } from '@/types/shared'

export default async function SettingsPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id: hash } = await params

  const result = await getWorkspaceData(hash)

  if (!result.success || !result.data) {
    notFound()
  }

  const { workspace } = result.data
  const namespace = workspace.metadata?.namespace || 'default'
  const name = workspace.metadata?.name || ''

  return (
    <div className="space-y-6">
      {/* General Information */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center gap-2">
            <Info className="h-4 w-4 text-muted-foreground" />
            <h3 className="text-sm font-semibold">General Information</h3>
          </div>
        </div>
        <div className="p-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Workspace Name</p>
              <p className="text-sm font-mono">{name}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Namespace</p>
              <p className="text-sm font-mono">{namespace}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Hash ID</p>
              <p className="text-sm font-mono">{workspace.status?.hash || hash}</p>
            </div>
            {workspace.metadata?.creationTimestamp && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Created</p>
                <p className="text-sm">
                  {new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Owner & Machine */}
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Owner</h3>
            </div>
          </div>
          <div className="p-4">
            <p className="text-sm font-medium">{workspace.spec?.ownedBy || 'Unknown'}</p>
            <p className="text-xs text-muted-foreground mt-1">
              This workspace belongs to this user
            </p>
          </div>
        </div>

        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <div className="flex items-center gap-2">
              <Server className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Work Machine</h3>
            </div>
          </div>
          <div className="p-4">
            <p className="text-sm font-mono">{workspace.spec?.workmachine || 'default'}</p>
            <p className="text-xs text-muted-foreground mt-1">
              The work machine running this workspace
            </p>
          </div>
        </div>
      </div>

      {/* Description */}
      <WorkspaceDescriptionForm
        workspaceName={name}
        namespace={namespace}
        currentDisplayName={workspace.spec?.displayName || ''}
      />

      {/* Danger Zone */}
      <WorkspaceDangerZone
        workspaceName={name}
        namespace={namespace}
        hash={hash}
      />
    </div>
  )
}
