'use client'

import { useSearchParams } from 'next/navigation'
import { AuthCard } from '@/components/auth/auth-card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, ArrowLeft } from 'lucide-react'
import { Link } from '@/components/ui/link'

const errorMessages: Record<string, string> = {
  Configuration: 'There is a problem with the server configuration. Check if the OAuth provider is properly configured.',
  AccessDenied: 'Access was denied. You may not have permission to sign in.',
  Verification: 'The verification token has expired or has already been used.',
  OAuthSignin: 'Error occurred while constructing an authorization URL.',
  OAuthCallback: 'Error occurred while handling the OAuth callback.',
  OAuthCreateAccount: 'Could not create OAuth provider user in the database.',
  EmailCreateAccount: 'Could not create email provider user in the database.',
  Callback: 'Error occurred in the OAuth callback handler route.',
  OAuthAccountNotLinked: 'This email is already associated with another account. Please sign in with the original provider.',
  EmailSignin: 'Check your email - we sent you a sign in link.',
  CredentialsSignin: 'Sign in failed. Check the details you provided are correct.',
  SessionRequired: 'Please sign in to access this page.',
  Default: 'An unexpected error occurred during authentication.',
}

export default function AuthErrorPage() {
  const searchParams = useSearchParams()
  const error = searchParams.get('error')
  const errorMessage = error ? errorMessages[error] || errorMessages.Default : errorMessages.Default

  return (
    <AuthCard
      title="Authentication Error"
      description="Something went wrong during authentication"
      icon={AlertCircle}
    >
      <div className="space-y-6">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {errorMessage}
          </AlertDescription>
        </Alert>

        {error === 'Configuration' && (
          <div className="space-y-2 text-sm text-muted-foreground">
            <p>Common causes:</p>
            <ul className="list-disc list-inside space-y-1">
              <li>Missing or incorrect OAuth client ID/secret</li>
              <li>Incorrect redirect URI configuration</li>
              <li>OAuth app not properly configured</li>
            </ul>
          </div>
        )}

        {error === 'OAuthAccountNotLinked' && (
          <div className="space-y-2 text-sm text-muted-foreground">
            <p>
              To link this provider to your existing account, please sign in with your original provider first, 
              then link additional providers from your account settings.
            </p>
          </div>
        )}

        <div className="flex flex-col gap-3">
          <Button asChild variant="default" size="auth">
            <Link href="/auth/login">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Sign In
            </Link>
          </Button>
          
          <Button asChild variant="outline" size="auth">
            <Link href="/">
              Go to Homepage
            </Link>
          </Button>
        </div>

        {process.env.NODE_ENV === 'development' && error && (
          <div className="mt-6 p-4 bg-muted rounded-lg">
            <p className="text-xs font-mono text-muted-foreground">
              Debug: error={error}
            </p>
          </div>
        )}
      </div>
    </AuthCard>
  )
}