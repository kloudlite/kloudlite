'use client'

import { useState } from 'react'
import { signIn } from 'next-auth/react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Building2, LogIn } from 'lucide-react'
import { cn } from '@/lib/utils'

interface SSOLoginProps {
  callbackUrl?: string
  className?: string
}

// Mock SSO provider mapping - in production this would come from backend
const SSO_PROVIDERS: Record<string, { provider: string; name: string }> = {
  'microsoft.com': { provider: 'azure-ad', name: 'Microsoft' },
  'outlook.com': { provider: 'azure-ad', name: 'Microsoft' },
  'google.com': { provider: 'google', name: 'Google' },
  'gmail.com': { provider: 'google', name: 'Google' },
  'github.com': { provider: 'github', name: 'GitHub' },
}

export function SSOLogin({ callbackUrl = '/teams', className }: SSOLoginProps) {
  const [email, setEmail] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSSOLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    
    if (!email) {
      setError('Please enter your email address')
      return
    }

    const domain = email.split('@')[1]
    if (!domain) {
      setError('Please enter a valid email address')
      return
    }

    const ssoProvider = SSO_PROVIDERS[domain.toLowerCase()]
    
    if (!ssoProvider) {
      setError(`SSO is not configured for ${domain}. Please contact your administrator.`)
      return
    }

    setIsLoading(true)
    try {
      await signIn(ssoProvider.provider, { callbackUrl })
      // Don't reset loading - let it persist during redirect
    } catch (error) {
      console.error('SSO login error:', error)
      setError('Failed to initiate SSO login. Please try again.')
      setIsLoading(false)
    }
  }

  return (
    <div className={cn("space-y-4", className)}>
      <form onSubmit={handleSSOLogin} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="sso-email">Work Email</Label>
          <Input
            id="sso-email"
            type="email"
            placeholder="you@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={isLoading}
            className="h-11"
          />
        </div>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <Button
          type="submit"
          disabled={isLoading}
          className="w-full h-11"
        >
          {isLoading ? (
            'Redirecting to SSO provider...'
          ) : (
            <>
              <Building2 className="h-4 w-4 mr-2" />
              Continue with SSO
            </>
          )}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-xs text-muted-foreground">
          Supported providers: Microsoft, Google, GitHub
        </p>
      </div>
    </div>
  )
}