import { LoginCard } from '@/components/console/login-card'
import { getRegistrationSession } from '@/lib/console-auth'
import { redirect } from 'next/navigation'

const ERROR_MESSAGES: Record<string, string> = {
  missing_params: 'Missing required parameters. Please try again.',
  invalid_state: 'Invalid state parameter. Please try again.',
  invalid_provider: 'Invalid OAuth provider. Please contact support.',
  oauth_exchange_failed: 'Failed to authenticate with the provider. Please try again.',
  no_email:
    'No email address was found in your account. Please ensure your account has a verified email.',
  access_denied: 'Access was denied. Please try again.',
}

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>
}) {
  // Check if user is already authenticated
  const session = await getRegistrationSession()

  // If authenticated, redirect to home
  if (session?.user) {
    redirect('/')
  }

  const { error } = await searchParams
  const errorMessage = error
    ? ERROR_MESSAGES[error] || 'An unexpected error occurred. Please try again.'
    : null

  // Not authenticated - show login page
  return (
    <div className="bg-background min-h-screen flex items-center justify-center p-4 sm:p-6 relative overflow-hidden">
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
      <LoginCard errorMessage={errorMessage} />

      {/* Bottom branding */}
      <div className="absolute bottom-6 left-0 right-0 text-center">
        <p className="text-muted-foreground/40 text-xs">
          Powered by Kloudlite · Cloud-Native Development Platform
        </p>
      </div>
    </div>
  )
}
