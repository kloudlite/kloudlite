'use client'

import { useEffect, useState } from 'react'

import { Loader2, CheckCircle, XCircle } from 'lucide-react'
import Link from 'next/link'
import { useSearchParams, useRouter } from 'next/navigation'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function PlatformInvitePage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const token = searchParams.get('token')
  
  const [status, setStatus] = useState<'loading' | 'success' | 'error' | 'invalid'>('loading')
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    if (!token) {
      setStatus('invalid')
      return
    }

    // Accept the invitation
    acceptInvitation(token)
  }, [token])

  const acceptInvitation = async (inviteToken: string) => {
    try {
      const response = await fetch('/api/platform-invite/accept', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ token: inviteToken }),
      })

      const data = await response.json()

      if (response.ok && data.success) {
        setStatus('success')
        // Redirect to login after 3 seconds
        setTimeout(() => {
          router.push('/auth/login')
        }, 3000)
      } else {
        setStatus('error')
        setErrorMessage(data.error || 'Failed to accept invitation')
      }
    } catch (error) {
      setStatus('error')
      setErrorMessage('An unexpected error occurred')
    }
  }

  return (
    <div className="flex min-h-[400px] items-center justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle>Platform Invitation</CardTitle>
        </CardHeader>
        <CardContent className="text-center">
          {status === 'loading' && (
            <div className="space-y-4 py-8">
              <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
              <p className="text-sm text-muted-foreground">Processing your invitation...</p>
            </div>
          )}
          
          {status === 'success' && (
            <div className="space-y-4 py-8">
              <CheckCircle className="h-12 w-12 mx-auto text-green-600" />
              <div className="space-y-2">
                <p className="font-medium">Invitation Accepted!</p>
                <p className="text-sm text-muted-foreground">
                  You&apos;ve been successfully added to the platform. Redirecting to login...
                </p>
              </div>
            </div>
          )}
          
          {status === 'error' && (
            <div className="space-y-4 py-8">
              <XCircle className="h-12 w-12 mx-auto text-red-600" />
              <div className="space-y-2">
                <p className="font-medium">Invitation Error</p>
                <p className="text-sm text-muted-foreground">{errorMessage}</p>
              </div>
              <div className="pt-4">
                <Link href="/auth/login">
                  <Button variant="outline">Go to Login</Button>
                </Link>
              </div>
            </div>
          )}
          
          {status === 'invalid' && (
            <div className="space-y-4 py-8">
              <XCircle className="h-12 w-12 mx-auto text-red-600" />
              <div className="space-y-2">
                <p className="font-medium">Invalid Invitation</p>
                <p className="text-sm text-muted-foreground">
                  This invitation link is invalid or has expired.
                </p>
              </div>
              <div className="pt-4">
                <Link href="/auth/login">
                  <Button variant="outline">Go to Login</Button>
                </Link>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}