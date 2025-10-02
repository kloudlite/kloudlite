import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { ProviderCard } from './provider-card'
import { getOAuthProviders } from './oauth-actions'
import type { OAuthProvider } from '@/lib/services/oauth-provider.service'

const PROVIDER_TYPES = ['google', 'github', 'microsoft'] as const

export default async function OAuthProvidersPage() {
  // Check authentication
  const session = await auth()
  if (!session || !session.user?.email) {
    redirect('/auth/signin')
  }

  // Check if user has admin or super-admin role
  const userRoles = session.user?.roles || []
  const hasAdminAccess = userRoles.includes('admin') || userRoles.includes('super-admin')
  const isSuperAdmin = userRoles.includes('super-admin')

  if (!hasAdminAccess) {
    redirect('/')
  }

  // Fetch OAuth providers
  let providers: Record<string, OAuthProvider> = {}
  let error: string | null = null

  try {
    providers = await getOAuthProviders()
  } catch (err: any) {
    console.error('Error fetching OAuth providers:', err)
    error = err.message || 'Failed to fetch OAuth providers'
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-gray-900">Unable to load OAuth providers</h2>
          <p className="mt-2 text-gray-600">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-7xl px-6 py-8 space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">OAuth Provider Configuration</h1>
        <p className="text-sm text-gray-600 mt-1.5">
          {isSuperAdmin ? 'Manage OAuth providers for user authentication' : 'View OAuth provider configurations'}
        </p>
      </div>

      {/* Providers Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {PROVIDER_TYPES.map((type) => {
          const provider = providers[type] || {
            type,
            enabled: false,
            clientId: '',
            clientSecret: '',
          }
          return (
            <ProviderCard
              key={type}
              provider={provider}
              displayName={type.charAt(0).toUpperCase() + type.slice(1)}
              isReadOnly={!isSuperAdmin}
            />
          )
        })}
      </div>
    </div>
  )
}