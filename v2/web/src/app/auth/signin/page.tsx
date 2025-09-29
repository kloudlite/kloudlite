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
    <div className="min-h-screen flex flex-col justify-center bg-gray-50">
      <div className="mx-auto w-full max-w-md px-6">
        {/* Logo */}
        <div className="flex justify-center mb-8">
          <KloudliteLogo className="h-10" linkToHome={false} />
        </div>

        {/* Sign In Card */}
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-200">
          {/* Header */}
          <div className="text-center mb-6">
            <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
              Sign in to your account
            </h1>
            <p className="mt-2 text-sm text-gray-600">
              Enter your credentials to access your dashboard
            </p>
          </div>

          <SignInForm enabledProviders={enabledProviders} />
        </div>
      </div>
    </div>
  )
}