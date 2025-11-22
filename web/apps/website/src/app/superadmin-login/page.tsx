'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Loader2, Shield, AlertCircle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@kloudlite/ui'
import { Alert, AlertDescription } from '@kloudlite/ui'

export default function SuperAdminLoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [status, setStatus] = useState<'validating' | 'success' | 'error'>('validating')
  const [errorMessage, setErrorMessage] = useState('')

  const validateToken = useCallback(async (token: string) => {
    try {
      // Call our Next.js API route which will validate with the Go API server
      // and create a NextAuth session cookie
      const response = await fetch('/api/superadmin-login/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ token }),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Token validation failed')
      }

      const data = await response.json()

      if (!data.success) {
        throw new Error('Invalid token')
      }

      // Session cookie is already set by the API route
      setStatus('success')

      // Redirect to dashboard after short delay
      setTimeout(() => {
        router.push('/')
      }, 1500)
    } catch (error) {
      console.error('Token validation error:', error)
      setStatus('error')
      setErrorMessage(
        error instanceof Error ? error.message : 'Failed to validate authentication token'
      )
    }
  }, [router])

  useEffect(() => {
    const token = searchParams.get('token')

    if (!token) {
      setStatus('error')
      setErrorMessage('Missing authentication token')
      return
    }

    // Validate token with API server
    validateToken(token)
  }, [searchParams, validateToken])

  return (
    <div className="bg-background flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="bg-primary/10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
            <Shield className="text-primary h-8 w-8" />
          </div>
          <CardTitle className="text-2xl">Super Admin Login</CardTitle>
          <CardDescription>
            {status === 'validating' && 'Validating your authentication token...'}
            {status === 'success' && 'Authentication successful!'}
            {status === 'error' && 'Authentication failed'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {status === 'validating' && (
            <div className="flex flex-col items-center gap-4 py-8">
              <Loader2 className="text-primary h-12 w-12 animate-spin" />
              <p className="text-muted-foreground text-sm">
                Please wait while we verify your credentials...
              </p>
            </div>
          )}

          {status === 'success' && (
            <div className="flex flex-col items-center gap-4 py-8">
              <div className="bg-success/10 text-success flex h-16 w-16 items-center justify-center rounded-full">
                <Shield className="h-8 w-8" />
              </div>
              <div className="text-center">
                <p className="text-success font-medium">Login Successful!</p>
                <p className="text-muted-foreground mt-1 text-sm">
                  Redirecting you to the dashboard...
                </p>
              </div>
            </div>
          )}

          {status === 'error' && (
            <div className="space-y-4">
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
              <div className="text-muted-foreground text-center text-sm">
                <p>This login link may have expired or is invalid.</p>
                <p className="mt-2">Please generate a new login URL from the console.</p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
