import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { apiClient } from '@/lib/api-client'
import { LogoutButton } from './components/logout-button'
import { ProviderCard } from './components/provider-card'

const PROVIDER_TYPES = ['google', 'github', 'microsoft'] as const

interface Provider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

export default async function SuperAdminDashboard() {
  // Check authentication
  const session = await auth()
  if (!session || !session.user?.roles?.includes('super-admin')) {
    redirect('/auth/signin')
  }

  // Fetch providers on server
  let providers: Provider[] = []
  let error: string | null = null

  try {
    const data = await apiClient.get<Record<string, Provider>>('/api/v1/providers')
    providers = Object.values(data || {})
  } catch (err) {
    console.error('Error fetching providers:', err)
    const errorObj = err instanceof Error ? err : new Error('Failed to fetch providers')
    error = errorObj.message
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-gray-900">Unable to load providers</h2>
          <p className="mt-2 text-gray-600">{error}</p>
        </div>
      </div>
    )
  }

  // Convert array to object for easier access
  const providersData = providers.reduce((acc, provider) => {
    acc[provider.type] = provider
    return acc
  }, {} as Record<string, Provider>)

  return (
    <div className="min-h-screen bg-white">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="flex items-center justify-between py-8 border-b border-gray-200">
          <div>
            <h1 className="text-2xl font-semibold text-gray-900">Provider Configuration</h1>
            <p className="mt-1 text-sm text-gray-600">Manage OAuth providers for authentication</p>
          </div>
          <LogoutButton />
        </div>

        {/* Providers Grid */}
        <div className="py-8">
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {PROVIDER_TYPES.map((type) => {
              const provider = providersData[type] || {
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
                />
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}