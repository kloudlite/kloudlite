'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
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
import { ForgotPasswordData } from '@/lib/auth/types'
import { forgotPasswordAction } from '@/actions/auth/forgot-password'
import { AlertCircle, CheckCircle2, ArrowLeft } from 'lucide-react'

export function ForgotPasswordForm() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const form = useForm<ForgotPasswordData>({
    defaultValues: {
      email: '',
    },
  })

  const onSubmit = async (data: ForgotPasswordData) => {
    setIsLoading(true)
    setError(null)
    setSuccess(null)

    try {
      const result = await forgotPasswordAction(data)
      
      if (result.success) {
        setSuccess(result.message || 'Password reset email sent successfully')
        form.reset()
      } else {
        setError(result.error || 'An error occurred')
      }
    } catch (err) {
      setError('An unexpected error occurred')
    } finally {
      setIsLoading(false)
    }
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

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
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
                    disabled={isLoading || !!success}
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
            disabled={isLoading || !!success}
          >
            {isLoading ? 'Sending reset link...' : 'Send reset link'}
          </Button>

          <div className="flex items-center justify-center">
            <Link href="/auth/login" className="flex items-center gap-2 text-sm">
              <ArrowLeft className="h-4 w-4" />
              Back to sign in
            </Link>
          </div>
        </form>
      </Form>
    </div>
  )
}