import Link from 'next/link'
import { SignInForm } from './signin-form'
import { apiClient } from '@/lib/api-client'

interface Provider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

async function getEnabledProviders() {
  try {
    const data = await apiClient.get<Record<string, Provider>>('/api/v1/providers')
    const providers = Object.values(data || {})
    return providers.filter(p => p.enabled && p.clientId)
  } catch (error) {
    console.error('Error fetching providers:', error)
    return []
  }
}

export default async function SignInPage() {
  const enabledProviders = await getEnabledProviders()

  return (
    <div className="min-h-screen flex flex-col justify-center bg-white">
      <div className="mx-auto w-full max-w-sm">
        {/* Logo/Brand */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-semibold tracking-tight text-gray-900">
            Sign in
          </h1>
          <p className="mt-2 text-sm text-gray-600">
            Sign in to your account to continue
          </p>
        </div>

        <SignInForm enabledProviders={enabledProviders} />
      </div>
    </div>
  )
}