'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { signIn } from 'next-auth/react'
import { Button, Input, Label, KloudliteLogo } from '@kloudlite/ui'
import { AlertCircle } from 'lucide-react'
import { cn } from '@kloudlite/lib'

interface Provider {
  type: string
  enabled: boolean
}

interface SignInFormProps {
  enabledProviders: Provider[]
}

// Cross marker component for grid
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

export function SignInForm({ enabledProviders }: SignInFormProps) {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loadingAction, setLoadingAction] = useState<string | null>(null)

  const loading = loadingAction !== null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoadingAction('credentials')

    try {
      const result = await signIn('credentials', {
        email,
        password,
        redirect: false,
      })

      if (result?.error) {
        setError('Invalid email or password')
        setLoadingAction(null)
        return
      }

      if (result?.ok) {
        router.push('/')
        router.refresh()
        return
      }

      // Unexpected response — reset so the user can retry
      setError('An error occurred. Please try again.')
      setLoadingAction(null)
    } catch (err) {
      console.error('Sign in error:', err)
      setError('An error occurred. Please try again.')
      setLoadingAction(null)
    }
  }

  const handleOAuthLogin = async (provider: string) => {
    setLoadingAction(provider)
    try {
      await signIn(provider, { callbackUrl: '/' })
    } catch (error) {
      console.error(`OAuth login error with ${provider}:`, error)
      setError(`Failed to sign in with ${provider}`)
    } finally {
      setLoadingAction(null)
    }
  }

  const hasEnabledProviders = enabledProviders.length > 0

  return (
    <>
      <style jsx>{`
        @keyframes pulseTopLeftToRight {
          0% { left: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { left: 100%; opacity: 0; }
        }
        @keyframes pulseRightTopToBottom {
          0% { top: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes pulseBottomRightToLeft {
          0% { right: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { right: 100%; opacity: 0; }
        }
        @keyframes pulseLeftBottomToTop {
          0% { bottom: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { bottom: 100%; opacity: 0; }
        }
        .pulse-top {
          animation: pulseTopLeftToRight 4s ease-in-out infinite;
        }
        .pulse-right {
          animation: pulseRightTopToBottom 4s ease-in-out infinite 1s;
        }
        .pulse-bottom {
          animation: pulseBottomRightToLeft 4s ease-in-out infinite 2s;
        }
        .pulse-left {
          animation: pulseLeftBottomToTop 4s ease-in-out infinite 3s;
        }
      `}</style>

      {/* Login card with animated border */}
      <div className="relative mx-auto w-full max-w-[480px] border border-border/80">
        {/* Animated border pulses */}
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          <div className="pulse-top absolute top-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
          <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
          <div className="pulse-bottom absolute bottom-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
          <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        </div>

        {/* Corner markers */}
        <div className="absolute inset-0 pointer-events-none">
          <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
          <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
          <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
          <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
        </div>

        {/* Content */}
        <div className="relative bg-background p-8 sm:p-12 lg:p-16 space-y-10">
          {/* Logo */}
          <div className="flex justify-center">
            <KloudliteLogo className="scale-125 transition-transform hover:scale-[1.35]" />
          </div>

          {/* Heading */}
          <div className="space-y-3 text-center">
            <h1 className="text-foreground text-3xl sm:text-4xl font-bold tracking-tight">
              Welcome back
            </h1>
            <p className="text-muted-foreground text-base">
              Sign in to your workspace dashboard
            </p>
          </div>

          {/* Error message */}
          {error && (
            <div className="border-destructive bg-destructive/10 text-destructive flex items-start gap-3 border p-4 text-sm">
              <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0" />
              <div>{error}</div>
            </div>
          )}

          {/* OAuth providers */}
          {hasEnabledProviders && (
            <div className="space-y-3">
              {enabledProviders.map((provider) => {
                const isProviderLoading = loadingAction === provider.type
                return (
                <Button
                  key={provider.type}
                  type="button"
                  variant="outline"
                  size="lg"
                  className="w-full gap-3 text-base font-medium transition-all duration-200 group"
                  disabled={loading}
                  onClick={() => handleOAuthLogin(provider.type)}
                >
                  {isProviderLoading ? (
                    <>
                      <svg
                        className="animate-spin h-5 w-5"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle
                          className="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          strokeWidth="4"
                        />
                        <path
                          className="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        />
                      </svg>
                      <span className="flex-1 text-left">Redirecting...</span>
                    </>
                  ) : (
                    <>
                  {provider.type === 'google' && (
                    <>
                      <svg className="h-5 w-5 transition-transform group-hover:scale-110" viewBox="0 0 24 24">
                        <path
                          d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                          fill="#4285F4"
                        />
                        <path
                          d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                          fill="#34A853"
                        />
                        <path
                          d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                          fill="#FBBC05"
                        />
                        <path
                          d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                          fill="#EA4335"
                        />
                      </svg>
                      <span className="flex-1 text-left">Continue with Google</span>
                    </>
                  )}
                  {provider.type === 'github' && (
                    <>
                      <svg className="h-5 w-5 transition-transform group-hover:scale-110" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          fillRule="evenodd"
                          d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
                          clipRule="evenodd"
                        />
                      </svg>
                      <span className="flex-1 text-left">Continue with GitHub</span>
                    </>
                  )}
                  {provider.type === 'microsoft' && (
                    <>
                      <svg className="h-5 w-5 transition-transform group-hover:scale-110" viewBox="0 0 21 21">
                        <path fill="#f25022" d="M0 0h10v10H0z" />
                        <path fill="#00a4ef" d="M11 0h10v10H11z" />
                        <path fill="#7fba00" d="M0 11h10v10H0z" />
                        <path fill="#ffb900" d="M11 11h10v10H11z" />
                      </svg>
                      <span className="flex-1 text-left">Continue with Microsoft</span>
                    </>
                  )}
                      <svg
                        className="h-4 w-4 opacity-80 group-hover:opacity-100 group-hover:translate-x-0.5 transition-all"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    </>
                  )}
                </Button>
                )
              })}
            </div>
          )}

          {/* Divider */}
          {hasEnabledProviders && (
            <div className="relative py-1">
              <div className="absolute inset-0 flex items-center">
                <span className="border-border/50 w-full border-t" />
              </div>
              <div className="relative flex justify-center">
                <span className="bg-background text-muted-foreground/70 px-4 text-xs font-medium uppercase tracking-wider">
                  Or continue with email
                </span>
              </div>
            </div>
          )}

          {/* Email/Password form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <Label htmlFor="email">Email address</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={loading}
              />
            </div>

            <div>
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={loading}
              />
            </div>

            <Button
              type="submit"
              size="lg"
              disabled={loading}
              className="w-full gap-3 text-base font-medium bg-primary text-primary-foreground hover:bg-primary/90 border-primary transition-all duration-200 group"
            >
              {loadingAction === 'credentials' ? (
                <>
                  <svg
                    className="animate-spin h-5 w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  <span className="flex-1 text-left">Signing in...</span>
                </>
              ) : (
                <>
                  <svg
                    className="h-5 w-5 transition-transform group-hover:scale-110"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"
                    />
                  </svg>
                  <span className="flex-1 text-left">Sign in</span>
                  <svg
                    className="h-4 w-4 opacity-80 group-hover:opacity-100 group-hover:translate-x-0.5 transition-all"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </>
              )}
            </Button>
          </form>
        </div>
      </div>
    </>
  )
}
