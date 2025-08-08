'use client'

import { useState } from "react"

import { Mail, RefreshCw } from "lucide-react"
import { useRouter } from "next/navigation"
import { useSession } from "next-auth/react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"



import { resendVerificationEmail } from "../verify-email/actions"

export default function EmailVerificationRequiredPage() {
  const { data: session } = useSession()
  const router = useRouter()
  const [resending, setResending] = useState(false)
  const [resendSuccess, setResendSuccess] = useState(false)

  const handleResend = async () => {
    if (!session?.user?.email) {return}
    
    setResending(true)
    setResendSuccess(false)
    
    try {
      const success = await resendVerificationEmail(session.user.email)
      if (success) {
        setResendSuccess(true)
      } else {
        // Error: Failed to resend verification email
      }
    } catch (_error) {
      // Error: Failed to resend verification email
    } finally {
      setResending(false)
    }
  }

  const handleCheckVerification = () => {
    // Force a session refresh by redirecting
    router.refresh()
    router.push('/overview')
  }

  return (
    <div className="container max-w-lg mx-auto flex h-screen w-full flex-col items-center justify-center">
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-orange-100">
            <Mail className="h-6 w-6 text-orange-600" />
          </div>
          <CardTitle className="text-xl">Email Verification Required</CardTitle>
          <CardDescription>
            Please verify your email address to access your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            <div className="text-center text-sm text-muted-foreground">
              <p>We&apos;ve sent a verification email to:</p>
              <p className="font-medium text-foreground mt-1">{session?.user?.email || 'your email'}</p>
            </div>

            <div className="space-y-4">
              {resendSuccess ? (
                <div className="rounded-lg bg-green-50 p-4 text-sm text-green-800">
                  Verification email has been sent! Please check your inbox.
                </div>
              ) : (
                <Button
                  variant="outline"
                  className="w-full"
                  onClick={handleResend}
                  disabled={resending}
                >
                  {resending ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                      Sending...
                    </>
                  ) : (
                    <>
                      <Mail className="mr-2 h-4 w-4" />
                      Resend Verification Email
                    </>
                  )}
                </Button>
              )}

              <Button
                className="w-full"
                onClick={handleCheckVerification}
              >
                I&apos;ve Verified My Email
              </Button>
            </div>

            <div className="text-center text-xs text-muted-foreground">
              <p>Once you&apos;ve verified your email, click the button above to continue.</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}