import { SignInForm } from './signin-form'
import { ThemeSwitcher } from '@kloudlite/ui'
import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { setThemeCookie } from '@/app/actions/theme'
import { getOAuthConfig } from '@/lib/oauth-config'

// Force dynamic rendering - this page reads OAuth config
export const dynamic = 'force-dynamic'

interface EnabledProvider {
  type: string
  enabled: boolean
}

function getEnabledProviders(): EnabledProvider[] {
  const config = getOAuthConfig()
  const providers: EnabledProvider[] = []

  if (config.google.enabled && config.google.clientId) {
    providers.push({ type: 'google', enabled: true })
  }
  if (config.github.enabled && config.github.clientId) {
    providers.push({ type: 'github', enabled: true })
  }
  if (config.microsoft.enabled && config.microsoft.clientId) {
    providers.push({ type: 'microsoft', enabled: true })
  }

  return providers
}

export default async function SignInPage() {
  // Check if user is already authenticated
  const session = await getSession()

  // Only redirect if user has a valid session with roles
  // (prevents redirect loop when OAuth creates a session without roles)
  if (session?.user) {
    const roles = session.user.roles || []
    const hasValidRole = roles.includes('user') || roles.includes('admin') || roles.includes('super-admin')
    if (hasValidRole) {
      redirect('/')
    }
  }

  const enabledProviders = getEnabledProviders()

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
