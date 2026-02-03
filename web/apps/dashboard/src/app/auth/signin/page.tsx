import { SignInForm } from './signin-form'
import { unauthenticatedApiClient } from '@/lib/api-client'
import { ThemeSwitcher } from '@kloudlite/ui'
import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { setThemeCookie } from '@/app/actions/theme'

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
    // External provider API not available - return empty array (OAuth providers configured in auth.ts)
    return []
  }
}

export default async function SignInPage() {
  // Check if user is already authenticated
  const session = await getSession()

  // If authenticated, redirect to home
  if (session?.user) {
    redirect('/')
  }

  const enabledProviders = await getEnabledProviders()

  return (
    <div className="bg-background min-h-screen flex items-center justify-center p-4 sm:p-6 relative overflow-hidden">
      {/* Theme switcher */}
      <div className="absolute top-6 right-6 z-10">
        <ThemeSwitcher setThemeCookie={setThemeCookie} />
      </div>

      {/* Grid pattern background */}
      <div className="absolute inset-0 pointer-events-none">
        {/* Vertical lines */}
        {[...Array(8)].map((_, i) => (
          <div
            key={`v-${i}`}
            className="absolute inset-y-0 w-px bg-foreground/5"
            style={{ left: `${(i + 1) * 12.5}%` }}
          />
        ))}
        {/* Horizontal lines */}
        {[...Array(8)].map((_, i) => (
          <div
            key={`h-${i}`}
            className="absolute inset-x-0 h-px bg-foreground/5"
            style={{ top: `${(i + 1) * 12.5}%` }}
          />
        ))}
      </div>

      {/* Login card */}
      <SignInForm enabledProviders={enabledProviders} />

      {/* Bottom branding */}
      <div className="absolute bottom-6 left-0 right-0 text-center">
        <p className="text-muted-foreground/40 text-xs">
          Powered by Kloudlite · Cloud Development Environments
        </p>
      </div>
    </div>
  )
}
