import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { Breadcrumb } from '@/components/breadcrumb'
import { EnvironmentNav } from '../_components/environment-nav'
import { getEnvironmentDetails } from '@/app/actions/environment.actions'
import { EnvironmentStatusIndicator } from '@/components/environment-status-indicator'
import { EnvironmentSnapshotsSheet } from '../_components/environment-snapshots-sheet'

interface LayoutProps {
  children: React.ReactNode
  params: Promise<{
    id: string
  }>
}

function formatTimeAgo(timestamp?: string): string {
  if (!timestamp) return 'Unknown'

  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
  if (diffDays < 30) {
    const weeks = Math.floor(diffDays / 7)
    return `${weeks} week${weeks > 1 ? 's' : ''} ago`
  }
  const months = Math.floor(diffDays / 30)
  return `${months} month${months > 1 ? 's' : ''} ago`
}

export default async function EnvironmentLayout({ children, params }: LayoutProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Await params (required in Next.js 15)
  const { id } = await params

  // Fetch real environment data using server action
  let environment
  const result = await getEnvironmentDetails(id)

  if (result.success && result.data) {
    const env = result.data.environment

    environment = {
      id,
      name: env.metadata!.name!,
      displayName: `${env.spec!.ownedBy || 'unknown'}/${env.metadata!.name}`,
      owner: env.spec!.ownedBy || 'unknown',
      status: env.status?.state || 'unknown',
      created: formatTimeAgo(env.metadata!.creationTimestamp),
    }
  } else {
    // Fallback to basic data if fetching fails
    environment = {
      id,
      name: id,
      displayName: id,
      owner: session.user?.email || 'unknown',
      status: 'unknown',
      created: 'Unknown',
    }
  }

  const breadcrumbItems = [
    { label: 'Environments', href: '/environments' },
    { label: environment.displayName },
  ]

  return (
    <>
      {/* Environment Header with Info */}
      <div className="bg-background border-b">
        <div className="mx-auto max-w-7xl px-6">
          {/* Breadcrumb */}
          <div className="py-4">
            <Breadcrumb items={breadcrumbItems} />
          </div>

          {/* Environment Header */}
          <div className="pb-4">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-2xl font-semibold">{environment.displayName}</h1>
                <div className="text-muted-foreground mt-1.5 flex items-center gap-4 text-sm">
                  <span>Owner: {environment.owner}</span>
                  <span>•</span>
                  <span>Created: {environment.created}</span>
                  <span>•</span>
                  <EnvironmentStatusIndicator
                    environmentName={environment.name}
                    initialState={environment.status}
                  />
                </div>
              </div>
              <EnvironmentSnapshotsSheet environmentName={environment.name} />
            </div>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <EnvironmentNav environmentId={id} />

      {/* Page Content */}
      <div className="flex-1">{children}</div>
    </>
  )
}
