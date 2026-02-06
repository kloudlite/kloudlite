import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { OAuthProvidersList } from './provider-card'
import { getOAuthProviders } from './oauth-actions'
import type { OAuthProvider } from '@/lib/services/oauth-provider.service'

export default async function OAuthProvidersPage() {
  const session = await getSession()
  if (!session || !session.user?.email) {
    redirect('/auth/signin')
  }

  const userRoles = session.user?.roles || []
  const hasAdminAccess = userRoles.includes('admin') || userRoles.includes('super-admin')
  const isSuperAdmin = userRoles.includes('super-admin')

  if (!hasAdminAccess) {
    redirect('/')
  }

  let providers: Record<string, OAuthProvider> = {}
  let error: string | null = null

  try {
    providers = await getOAuthProviders()
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to fetch OAuth providers'
  }

  return (
    <div className="mx-auto max-w-7xl space-y-6 px-6 py-8">
      {/* Page Header */}
      <div>
        <h1 className="text-foreground text-2xl font-semibold">OAuth Provider Configuration</h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          {isSuperAdmin
            ? 'Configure OAuth providers to enable third-party authentication for your users'
            : 'View OAuth provider configurations'}
        </p>
      </div>

      {/* Error State */}
      {error && (
        <div className="bg-destructive/10 border-destructive/20 rounded-md border p-4">
          <p className="text-destructive text-sm">{error}</p>
        </div>
      )}

      {/* Providers */}
      <OAuthProvidersList providers={providers} isReadOnly={!isSuperAdmin} />
    </div>
  )
}
