'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { signIn } from 'next-auth/react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface Provider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

interface SignInFormProps {
  enabledProviders: Provider[]
}

export function SignInForm({ enabledProviders }: SignInFormProps) {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const result = await signIn('credentials', {
        email,
        password,
        redirect: false,
      })

      if (result?.error) {
        setError('Invalid email or password')
        setLoading(false)
        return
      }

      if (result?.ok) {
        router.push('/')
        router.refresh()
      }
    } catch (err) {
      console.error('Sign in error:', err)
      setError('An error occurred. Please try again.')
      setLoading(false)
    }
  }

  const handleOAuthLogin = async (provider: string) => {
    setLoading(true)
    try {
      await signIn(provider, { callbackUrl: '/' })
    } catch (error) {
      console.error(`OAuth login error with ${provider}:`, error)
      setError(`Failed to sign in with ${provider}`)
      setLoading(false)
    }
  }

  const hasEnabledProviders = enabledProviders.length > 0

  return (
    <form onSubmit={handleSubmit}>
      <div className="space-y-5">
        {error && (
          <Alert variant="destructive" className="py-3">
            <AlertDescription className="text-sm">{error}</AlertDescription>
          </Alert>
        )}

        <div className="space-y-2">
          <Label htmlFor="email" className="text-sm font-medium text-foreground">
            Email
          </Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="name@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={loading}
            className="h-10 px-3 text-sm"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="password" className="text-sm font-medium text-foreground">
            Password
          </Label>
          <Input
            id="password"
            type="password"
            autoComplete="current-password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={loading}
            className="h-10 px-3 text-sm"
          />
        </div>

        <Button
          type="submit"
          className="h-10 w-full text-sm font-medium"
          disabled={loading}
        >
          {loading ? 'Signing in...' : 'Sign in'}
        </Button>

        {/* OAuth providers section - only show if there are enabled providers */}
        {hasEnabledProviders && (
          <>
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs">
                <span className="bg-card px-2 text-muted-foreground">Or continue with</span>
              </div>
            </div>

            <div className="space-y-3">
              {enabledProviders.map((provider) => (
                <Button
                  key={provider.type}
                  type="button"
                  variant="outline"
                  className="h-10 w-full gap-2 text-sm font-medium"
                  disabled={loading}
                  onClick={() => handleOAuthLogin(provider.type)}
                >
                  {provider.type === 'google' && (
                    <>
                      <svg className="h-5 w-5" viewBox="0 0 24 24">
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
                      Continue with Google
                    </>
                  )}
                  {provider.type === 'github' && (
                    <>
                      <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          fillRule="evenodd"
                          d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
                          clipRule="evenodd"
                        />
                      </svg>
                      Continue with GitHub
                    </>
                  )}
                  {provider.type === 'microsoft' && (
                    <>
                      <svg className="h-5 w-5" viewBox="0 0 21 21">
                        <path fill="#f25022" d="M0 0h10v10H0z" />
                        <path fill="#00a4ef" d="M11 0h10v10H11z" />
                        <path fill="#7fba00" d="M0 11h10v10H0z" />
                        <path fill="#ffb900" d="M11 11h10v10H11z" />
                      </svg>
                      Continue with Microsoft
                    </>
                  )}
                </Button>
              ))}
            </div>
          </>
        )}
      </div>
    </form>
  )
}
