import { SignInForm } from './signin-form'
import { unauthenticatedApiClient } from '@/lib/api-client'
import { KloudliteLogo } from '@/components/kloudlite-logo'

// Force dynamic rendering - this page fetches providers from API
export const dynamic = 'force-dynamic'

interface Provider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

async function getEnabledProviders() {
  try {
    const data = await unauthenticatedApiClient.get<Record<string, Provider>>('/api/v1/providers')
    const providers = Object.values(data || {})
    return providers.filter((p) => p.enabled)
  } catch (error) {
    console.error('Error fetching providers:', error)
    return []
  }
}

export default async function SignInPage() {
  const enabledProviders = await getEnabledProviders()

  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      {/* Left side - Branding */}
      <div className="hidden flex-col justify-between bg-gray-900 p-12 text-white lg:flex">
        <div>
          <KloudliteLogo className="h-8" variant="white" linkToHome={false} />
        </div>
        <div>
          <h2 className="mb-4 text-2xl font-light">Cloud Development Environments</h2>
          <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
            Designed to reduce the development loop
          </p>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="bg-card flex items-center justify-center p-8">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="mb-8 lg:hidden">
            <KloudliteLogo className="h-8" linkToHome={false} />
          </div>

          <div className="mb-8">
            <h1 className="text-foreground text-lg font-medium">Sign in</h1>
            <p className="text-muted-foreground mt-1 text-sm">Access your workspace</p>
          </div>

          <SignInForm enabledProviders={enabledProviders} />
        </div>
      </div>
    </div>
  )
}
