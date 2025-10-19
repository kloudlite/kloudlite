import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { connectionTokenService } from '@/lib/services/connection-token.service'
import { ConnectionTokensList } from './_components/connection-tokens-list'

export default async function ConnectionTokensPage() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  // Fetch connection tokens from API
  let tokens = []
  let error = null

  try {
    tokens = await connectionTokenService.listTokens()
  } catch (err) {
    console.error('Failed to fetch connection tokens:', err)
    error = err instanceof Error ? err.message : 'Failed to fetch connection tokens'
    tokens = []
  }

  return (
    <div className="mx-auto max-w-7xl px-6 py-8 space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-semibold text-foreground">Connection Tokens</h1>
        <p className="text-sm text-muted-foreground mt-1.5">
          Manage API tokens for accessing Kloudlite workspaces from external applications
        </p>
      </div>

      {/* Error Display */}
      {error && (
        <div className="rounded-md bg-destructive/10 border border-destructive/20 p-4">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-destructive">
                Failed to load connection tokens
              </h3>
              <div className="mt-2 text-sm text-destructive/80">
                <p>{error}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Connection Tokens List */}
      <ConnectionTokensList tokens={tokens} />
    </div>
  )
}
