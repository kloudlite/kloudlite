import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getSession } from '@/lib/get-session'
import { EnvironmentNav } from '../_components/environment-nav'
import { getEnvironmentByHash } from '@/app/actions/environment.actions'
import { EnvironmentStatusIndicator } from '@/components/environment-status-indicator'
import { EnvironmentSnapshotsSheet } from '../_components/environment-snapshots-sheet'
import { EnvironmentCompositionButton } from '../_components/environment-composition-button'
import { ArrowLeft } from 'lucide-react'

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
  // id is now the environment hash (8-char hex)
  const { id: hash } = await params

  // Fetch real environment data using server action
  let environment
  const result = await getEnvironmentByHash(hash)

  if (result.success && result.data) {
    const env = result.data.environment

    environment = {
      hash,
      name: env.metadata!.name!,
      displayName: `${env.spec!.ownedBy || 'unknown'}/${env.metadata!.name}`,
      owner: env.spec!.ownedBy || 'unknown',
      status: env.status?.state || 'unknown',
      created: formatTimeAgo(env.metadata!.creationTimestamp),
    }
  } else {
    // Fallback to basic data if fetching fails
    environment = {
      hash,
      name: hash,
      displayName: hash,
      owner: session.user?.email || 'unknown',
      status: 'unknown',
      created: 'Unknown',
    }
  }

  return (
    <>
      {/* Back button */}
      <div className="mb-8">
        <Link
          href="/environments"
          className="group inline-flex items-center gap-2 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm"
        >
          <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
          <span className="relative">
            Back to Environments
            <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
          </span>
        </Link>
      </div>

      {/* Environment Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between gap-4 mb-2">
          <h1 className="text-2xl font-semibold tracking-tight truncate">{environment.displayName}</h1>
          <div className="flex-shrink-0 flex items-center gap-2">
            <EnvironmentCompositionButton environmentName={environment.name} />
            <EnvironmentSnapshotsSheet environmentName={environment.name} />
          </div>
        </div>
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <span>{environment.owner}</span>
          <span>•</span>
          <span>{environment.created}</span>
          <span>•</span>
          <EnvironmentStatusIndicator
            environmentName={environment.name}
            initialState={environment.status}
          />
        </div>
      </div>

      {/* Navigation */}
      <EnvironmentNav environmentId={hash} />

      {/* Page Content */}
      <div className="flex-1">{children}</div>
    </>
  )
}
