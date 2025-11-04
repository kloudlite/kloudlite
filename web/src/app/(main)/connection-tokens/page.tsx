import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import {
  connectionTokenService,
  type ConnectionToken,
} from '@/lib/services/connection-token.service'
import { ConnectionTokensList } from './_components/connection-tokens-list'

export default async function ConnectionTokensPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Fetch connection tokens from API
  let tokens: ConnectionToken[] = []
  let error: string | null = null

  try {
    tokens = await connectionTokenService.listTokens()
  } catch (err) {
    console.error('Failed to fetch connection tokens:', err)
    error = err instanceof Error ? err.message : 'Failed to fetch connection tokens'
    tokens = []
  }

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Title and Filter Section */}
      <div className="mb-8">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold">Connection Tokens</h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            Manage API tokens for accessing Kloudlite workspaces from external applications
          </p>
        </div>

        {/* Error Display */}
        {error && (
          <div className="mb-6 rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20">
            <div className="flex">
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                  Failed to load connection tokens
                </h3>
                <div className="mt-2 text-sm text-red-700 dark:text-red-300">
                  <p>{error}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Connection Tokens List */}
        <ConnectionTokensList tokens={tokens} />
      </div>
    </main>
  )
}
