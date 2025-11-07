import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { Breadcrumb } from '@/components/breadcrumb'
import { EnvironmentNav } from '../_components/environment-nav'
import { environmentService } from '@/lib/services/environment.service'

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

  // Fetch real environment data
  let environment
  try {
    const env = await environmentService.getEnvironment(id)

    // Use the createdBy field which contains the full email, then extract username
    const ownerEmail = env.spec.createdBy || 'unknown'
    const owner = ownerEmail.includes('@') ? ownerEmail.split('@')[0] : ownerEmail

    environment = {
      id,
      name: env.metadata.name,
      owner,
      status: env.status?.state || 'unknown',
      created: formatTimeAgo(env.metadata.creationTimestamp),
    }
  } catch (error) {
    console.error('Failed to fetch environment:', error)
    // Fallback to basic data if API fails
    environment = {
      id,
      name: id,
      owner: session.user?.email || 'unknown',
      status: 'unknown',
      created: 'Unknown',
    }
  }

  const breadcrumbItems = [
    { label: 'Environments', href: '/environments' },
    { label: environment.name },
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
                <h1 className="text-2xl font-semibold">{environment.name}</h1>
                <div className="text-muted-foreground mt-1.5 flex items-center gap-4 text-sm">
                  <span>Owner: {environment.owner}</span>
                  <span>•</span>
                  <span>Created: {environment.created}</span>
                  <span>•</span>
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      environment.status === 'ready' || environment.status === 'active'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                        : environment.status === 'creating' || environment.status === 'updating'
                          ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                          : environment.status === 'error' || environment.status === 'failed'
                            ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                            : environment.status === 'deleting'
                              ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                              : 'bg-secondary text-secondary-foreground'
                    }`}
                  >
                    {environment.status}
                  </span>
                </div>
              </div>
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
