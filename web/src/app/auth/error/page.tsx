'use client'

import { useSearchParams } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import Link from 'next/link'

export default function AuthErrorPage() {
  const searchParams = useSearchParams()
  const error = searchParams.get('error')
  const customMessage = searchParams.get('message')

  const errorMessages: Record<string, string> = {
    Configuration: 'There is a problem with the server configuration.',
    AccessDenied: 'You do not have permission to sign in.',
    Verification: 'The verification token has expired or has already been used.',
    Default: 'An error occurred during authentication.',
  }

  // Use custom message if provided, otherwise use default error message
  const errorMessage = customMessage
    ? decodeURIComponent(customMessage)
    : errorMessages[error || 'Default'] || errorMessages.Default

  // Check if user is not registered
  const isNotRegistered = errorMessage.includes('not registered')

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle>{isNotRegistered ? 'Account Not Found' : 'Authentication Error'}</CardTitle>
          <CardDescription className={isNotRegistered ? 'text-info' : 'text-destructive'}>
            {errorMessage}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {isNotRegistered && (
            <div className="border-border bg-muted text-foreground rounded-md border p-3 text-sm">
              <p className="font-semibold">Next steps:</p>
              <ul className="mt-1 list-inside list-disc space-y-1">
                <li>Contact your system administrator</li>
                <li>Request account creation with your email address</li>
                <li>Once created, you can sign in with any OAuth provider</li>
              </ul>
            </div>
          )}
          <Button asChild className="w-full">
            <Link href="/auth/signin">Back to Sign In</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
