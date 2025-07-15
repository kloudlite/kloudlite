'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useForm } from 'react-hook-form'
import { verifyEmailAction, resendVerificationEmailAction } from '@/actions/auth/verify-email'
import { AlertCircle, CheckCircle2, Mail, Loader2 } from 'lucide-react'

interface ResendFormData {
  email: string
}

export function VerifyEmail() {
  const searchParams = useSearchParams()
  const token = searchParams.get('token')
  const [isVerifying, setIsVerifying] = useState(false)
  const [isResending, setIsResending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [showResendForm, setShowResendForm] = useState(false)

  const form = useForm<ResendFormData>({
    defaultValues: {
      email: '',
    },
  })

  useEffect(() => {
    if (token) {
      verifyEmail(token)
    } else {
      setError('No verification token provided')
      setShowResendForm(true)
    }
  }, [token])

  const verifyEmail = async (verificationToken: string) => {
    setIsVerifying(true)
    setError(null)

    try {
      const result = await verifyEmailAction(verificationToken)
      
      if (!result.success) {
        setError(result.error || 'Verification failed')
        setShowResendForm(true)
      }
    } catch (err) {
      setError('An unexpected error occurred')
      setShowResendForm(true)
    } finally {
      setIsVerifying(false)
    }
  }

  const onResendSubmit = async (data: ResendFormData) => {
    setIsResending(true)
    setError(null)
    setSuccess(null)

    try {
      const result = await resendVerificationEmailAction(data.email)
      
      if (result.success) {
        setSuccess(result.message || 'Verification email sent!')
        form.reset()
      } else {
        setError(result.error || 'Failed to send verification email')
      }
    } catch (err) {
      setError('An unexpected error occurred')
    } finally {
      setIsResending(false)
    }
  }

  if (isVerifying) {
    return (
      <div className="text-center space-y-4">
        <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
        <p className="text-muted-foreground">Verifying your email address...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {success && (
        <Alert variant="success">
          <CheckCircle2 className="h-4 w-4" />
          <AlertDescription>{success}</AlertDescription>
        </Alert>
      )}

      {showResendForm && (
        <>
          <div className="text-center space-y-2">
            <Mail className="h-12 w-12 mx-auto text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              Enter your email address to receive a new verification link
            </p>
          </div>

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onResendSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="email"
                rules={{
                  required: 'Email is required',
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: 'Invalid email address',
                  },
                }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input
                        type="email"
                        placeholder="name@example.com"
                        autoComplete="email"
                        disabled={isResending}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <Button 
                type="submit" 
                className="w-full" 
                size="auth" 
                disabled={isResending}
              >
                {isResending ? 'Sending...' : 'Send verification email'}
              </Button>
            </form>
          </Form>
        </>
      )}

      <div className="text-center">
        <Link href="/auth/login" className="text-sm">
          Back to sign in
        </Link>
      </div>
    </div>
  )
}