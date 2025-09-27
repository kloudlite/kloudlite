'use client'

import { useEffect, useState } from "react"

import { Mail, CheckCircle, XCircle, Loader2 } from "lucide-react"
import Link from "next/link"
import { useSearchParams, useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { verifyEmail, resendVerificationEmail } from "./actions"

export default function VerifyEmailPage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const token = searchParams.get('token')
  const email = searchParams.get('email') || 'm@example.com'
  
  const [verificationStatus, setVerificationStatus] = useState<'pending' | 'verifying' | 'success' | 'error'>('pending')
  const [resending, setResending] = useState(false)

  useEffect(() => {
    if (token) {
      // Automatically verify when token is present
      setVerificationStatus('verifying')
      handleVerification()
    }
  }, [token])

  const handleVerification = async () => {
    if (!token) {return}
    
    try {
      const result = await verifyEmail(token)
      if (result.success) {
        setVerificationStatus('success')
      } else {
        setVerificationStatus('error')
      }
    } catch (_error) {
      setVerificationStatus('error')
    }
  }

  const handleResend = async () => {
    setResending(true)
    try {
      const success = await resendVerificationEmail(email)
      if (success) {
        // Show success message
        // TODO: Add toast notification
      }
    } catch (_error) {
      // Failed to resend email
    } finally {
      setResending(false)
    }
  }

  // If we have a token, show verification in progress or result
  if (token) {
    if (verificationStatus === 'verifying') {
      return (
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
              <Loader2 className="h-6 w-6 text-primary animate-spin" />
            </div>
            <CardTitle className="text-xl">Verifying your email...</CardTitle>
            <CardDescription>
              Please wait while we confirm your email address
            </CardDescription>
          </CardHeader>
        </Card>
      )
    }

    if (verificationStatus === 'success') {
      return (
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-green-100">
              <CheckCircle className="h-6 w-6 text-green-600" />
            </div>
            <CardTitle className="text-xl">Email verified!</CardTitle>
            <CardDescription>
              Your email has been successfully verified. You can now log in to your account.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button 
              className="w-full" 
              onClick={() => router.push('/auth/login?verified=true')}
            >
              Continue to login
            </Button>
          </CardContent>
        </Card>
      )
    }

    if (verificationStatus === 'error') {
      return (
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-red-100">
              <XCircle className="h-6 w-6 text-red-600" />
            </div>
            <CardTitle className="text-xl">Verification failed</CardTitle>
            <CardDescription>
              The verification link may have expired or is invalid
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Button 
                onClick={handleResend} 
                disabled={resending}
                className="w-full"
              >
                {resending ? 'Sending...' : 'Resend verification email'}
              </Button>
              <Button 
                variant="outline" 
                className="w-full"
                onClick={() => router.push('/auth/login')}
              >
                Back to login
              </Button>
            </div>
          </CardContent>
        </Card>
      )
    }
  }

  // Default state - waiting for email (when no token in URL)
  return (
    <Card>
      <CardHeader className="text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
          <Mail className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-xl">Check your email</CardTitle>
        <CardDescription>
          We sent a verification link to your email address
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid gap-6">
          <div className="text-center text-sm text-muted-foreground">
            <p>We've sent a verification email to:</p>
            <p className="font-medium text-foreground mt-1">{email}</p>
          </div>
          
          <div className="space-y-4">
            <Button variant="outline" className="w-full">
              Open email app
            </Button>
            
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">
                  Didn't receive email?
                </span>
              </div>
            </div>
            
            <Button 
              variant="link" 
              className="w-full"
              onClick={handleResend}
              disabled={resending}
            >
              {resending ? 'Sending...' : 'Click to resend'}
            </Button>
          </div>
          
          <Link 
            href="/auth/login" 
            className="text-center text-sm text-muted-foreground hover:text-primary transition-colors"
          >
            Skip, I'll confirm later
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}