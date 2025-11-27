import { KloudliteLogo } from '@/components/kloudlite-logo'
import { OAuthButtons } from '@/components/console/oauth-buttons'
import { getRegistrationSession } from '@/lib/console-auth'
import { redirect } from 'next/navigation'
import { AlertCircle } from 'lucide-react'

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
    <div className="grid min-h-screen lg:grid-cols-2">
      {/* Left side - Branding */}
      <div className="hidden flex-col justify-between overflow-hidden bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 p-12 text-white lg:flex">
        <div className="py-4">
          <KloudliteLogo variant="white" className="origin-left scale-150" />
        </div>
        <div className="space-y-6">
          <div>
            <h2 className="mb-4 text-3xl font-semibold">Build Faster, Ship Smarter</h2>
            <p className="max-w-md text-base leading-relaxed text-gray-300">
              Spin up production-ready development environments in seconds. Focus on building, not
              configuring.
            </p>
          </div>
          <div className="space-y-3 text-sm text-gray-400">
            <div className="flex items-start gap-3">
              <svg
                className="mt-1 h-5 w-5 flex-shrink-0 text-green-400"
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
              <span>Ready in seconds, not hours</span>
            </div>
            <div className="flex items-start gap-3">
              <svg
                className="mt-1 h-5 w-5 flex-shrink-0 text-green-400"
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
              <span>Your cloud, your control</span>
            </div>
            <div className="flex items-start gap-3">
              <svg
                className="mt-1 h-5 w-5 flex-shrink-0 text-green-400"
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
              <span>Zero vendor lock-in, 100% open source</span>
            </div>
          </div>
        </div>
      </div>

      {/* Right side - Form */}
      <div className="bg-background flex items-center justify-center p-8">
        <div className="w-full max-w-md">
          {/* Mobile logo */}
          <div className="mb-12 flex justify-center lg:hidden">
            <KloudliteLogo />
          </div>

          <div className="mb-10 text-center">
            <h1 className="text-foreground mb-3 text-3xl font-bold tracking-tight">
              Welcome to Kloudlite
            </h1>
            <p className="text-muted-foreground text-base">Sign in to access your installations</p>
          </div>

          <div className="space-y-8">
            {errorMessage && (
              <div className="border-destructive/50 bg-destructive/10 text-destructive flex items-start gap-3 rounded-lg border p-4 text-sm">
                <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0" />
                <div>{errorMessage}</div>
              </div>
            )}

            <div className="space-y-3">
              <OAuthButtons />
            </div>

            <div className="relative py-4">
              <div className="absolute inset-0 flex items-center">
                <span className="border-border w-full border-t" />
              </div>
              <div className="relative flex justify-center">
                <span className="bg-background text-muted-foreground px-4 text-sm font-medium tracking-wider uppercase">
                  New to Kloudlite?
                </span>
              </div>
            </div>

            <div className="space-y-4 text-center">
              <p className="text-muted-foreground text-sm leading-relaxed">
                Get started with a free account and deploy your first environment in minutes.
              </p>
              <div className="text-muted-foreground flex items-center justify-center gap-6 text-xs">
                <div className="flex items-center gap-1.5">
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
                  <span>No credit card required</span>
                </div>
                <div className="flex items-center gap-1.5">
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
          </div>
        </div>
      </div>
    </div>
  )
}
