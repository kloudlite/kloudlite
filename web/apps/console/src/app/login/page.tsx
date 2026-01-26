import { KloudliteLogo } from '@/components/kloudlite-logo'
import { OAuthButtons } from '@/components/console/oauth-buttons'
import { getRegistrationSession } from '@/lib/console-auth'
import { redirect } from 'next/navigation'
import { AlertCircle } from 'lucide-react'
import { cn } from '@kloudlite/lib'

// Cross marker component for grid
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

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
    <div className="bg-background min-h-screen flex items-center justify-center p-6 relative overflow-hidden">
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

      {/* Login card with grid border */}
      <div className="relative mx-auto w-full max-w-xl border border-border">
        {/* Corner markers */}
        <div className="absolute inset-0 pointer-events-none">
          <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
          <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
          <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
          <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
        </div>

        {/* Content */}
        <div className="relative bg-background p-10 lg:p-16 space-y-10">
          {/* Logo */}
          <div className="flex justify-center">
            <KloudliteLogo className="scale-125" />
          </div>

          {/* Heading */}
          <div className="space-y-3 text-center">
            <h1 className="text-foreground text-4xl font-bold tracking-tight">
              Welcome to Kloudlite
            </h1>
            <p className="text-muted-foreground text-lg">
              Sign in to manage your cloud installations
            </p>
          </div>

          {/* Error message */}
          {errorMessage && (
            <div className="border-destructive bg-destructive/10 text-destructive flex items-start gap-3 border p-4 text-base">
              <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0" />
              <div>{errorMessage}</div>
            </div>
          )}

          {/* OAuth buttons */}
          <div className="space-y-4">
            <OAuthButtons />
          </div>

          {/* Divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="border-border w-full border-t" />
            </div>
            <div className="relative flex justify-center">
              <span className="bg-background text-muted-foreground px-4 text-sm font-medium uppercase">
                New to Kloudlite?
              </span>
            </div>
          </div>

          {/* Bottom section */}
          <div className="space-y-6 text-center">
            <p className="text-muted-foreground text-base leading-relaxed">
              Get started with a free account and deploy your first environment in minutes.
            </p>
            <div className="flex items-center justify-center gap-8 text-muted-foreground text-sm">
              <div className="flex items-center gap-2">
                <svg
                  className="text-success h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span>No credit card</span>
              </div>
              <div className="flex items-center gap-2">
                <svg
                  className="text-success h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span>Free forever</span>
              </div>
            </div>
          </div>

          {/* Development backdoor */}
          {process.env.NODE_ENV !== 'production' && (
            <div className="border-t border-border pt-6">
              <a
                href="/api/dev-login"
                className="text-muted-foreground hover:text-foreground text-xs text-center block"
              >
                [Dev] Quick login as karthik@kloudlite.io
              </a>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
