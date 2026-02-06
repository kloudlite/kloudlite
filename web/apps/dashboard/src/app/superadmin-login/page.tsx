'use client'

import { useEffect, useState, useCallback } from 'react'
import { useSearchParams } from 'next/navigation'
import { signIn, signOut } from 'next-auth/react'
import { Loader2, Shield, AlertCircle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, Alert, AlertDescription } from '@kloudlite/ui'

export default function SuperAdminLoginPage() {
  const searchParams = useSearchParams()
  const [status, setStatus] = useState<'validating' | 'signing-out' | 'signing-in' | 'success' | 'error'>('validating')
  const [errorMessage, setErrorMessage] = useState('')

  const validateToken = useCallback(async (token: string) => {
    try {
      // Step 1: Sign out any existing session first
      setStatus('signing-out')
      await signOut({ redirect: false })

      // Step 2: Sign in with the super-admin token via NextAuth credentials provider
      setStatus('signing-in')
      const result = await signIn('credentials', {
        superadminToken: token,
        redirect: false,
      })

      if (result?.error) {
        throw new Error(result.error === 'CredentialsSignin'
          ? 'Invalid or expired super-admin token'
          : result.error)
      }

      if (!result?.ok) {
        throw new Error('Failed to sign in')
      }

      // Success - NextAuth has set the session cookie
      setStatus('success')

      // Redirect to admin dashboard immediately
      window.location.href = '/admin'
    } catch (error) {
      console.error('Token validation error:', error)
      setStatus('error')
      setErrorMessage(
        error instanceof Error ? error.message : 'Failed to validate authentication token'
      )
    }
  }, [])

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
            {status === 'validating' && 'Preparing authentication...'}
            {status === 'signing-out' && 'Signing out existing session...'}
            {status === 'signing-in' && 'Signing in as super admin...'}
            {status === 'success' && 'Authentication successful!'}
            {status === 'error' && 'Authentication failed'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {(status === 'validating' || status === 'signing-out' || status === 'signing-in') && (
            <div className="flex flex-col items-center gap-4 py-8">
              <Loader2 className="text-primary h-12 w-12 animate-spin" />
              <p className="text-muted-foreground text-sm">
                {status === 'validating' && 'Please wait while we verify your credentials...'}
                {status === 'signing-out' && 'Clearing existing session...'}
                {status === 'signing-in' && 'Establishing super admin session...'}
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
                  Redirecting you to the admin dashboard...
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
