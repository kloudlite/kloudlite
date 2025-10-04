import Link from 'next/link'
import { SignInForm } from './signin-form'
import { unauthenticatedApiClient } from '@/lib/api-client'
import { KloudliteLogo } from '@/components/kloudlite-logo'

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
    return providers.filter(p => p.enabled)
  } catch (error) {
    console.error('Error fetching providers:', error)
    return []
  }
}

export default async function SignInPage() {
  const enabledProviders = await getEnabledProviders()

  return (
    <div className="min-h-screen grid lg:grid-cols-2">
      {/* Left side - Branding */}
      <div className="hidden lg:flex bg-gray-900 text-white p-12 flex-col justify-between">
        <div>
          <KloudliteLogo className="h-8" variant="white" linkToHome={false} />
        </div>
        <div>
          <h2 className="text-2xl font-light mb-4">Cloud Development Environments</h2>
          <p className="text-sm text-gray-400 leading-relaxed max-w-md">
            Designed to reduce the development loop
          </p>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="flex items-center justify-center p-8 bg-white">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="lg:hidden mb-8">
            <KloudliteLogo className="h-8" linkToHome={false} />
          </div>

          <div className="mb-8">
            <h1 className="text-lg font-medium text-gray-900">
              Sign in
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Access your workspace
            </p>
          </div>

          <SignInForm enabledProviders={enabledProviders} />
        </div>
      </div>
    </div>
  )
}