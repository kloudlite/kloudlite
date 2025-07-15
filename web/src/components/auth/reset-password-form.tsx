'use client'

import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useSearchParams } from 'next/navigation'
import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { PasswordInput } from './password-input'
import { ResetPasswordData } from '@/lib/auth/types'
import { resetPasswordAction } from '@/actions/auth/reset-password'
import { AlertCircle, ArrowLeft } from 'lucide-react'

interface ResetPasswordFormData {
  password: string
  confirmPassword: string
}

export function ResetPasswordForm() {
  const searchParams = useSearchParams()
  const token = searchParams.get('token')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<ResetPasswordFormData>({
    defaultValues: {
      password: '',
      confirmPassword: '',
    },
  })

  const password = form.watch('password')

  useEffect(() => {
    if (!token) {
      setError('Invalid reset link. Please request a new password reset.')
    }
  }, [token])

  const onSubmit = async (data: ResetPasswordFormData) => {
    if (!token) return

    setIsLoading(true)
    setError(null)

    try {
      const resetData: ResetPasswordData = {
        token,
        password: data.password,
        confirmPassword: data.confirmPassword,
      }

      const result = await resetPasswordAction(resetData)
      
      if (!result.success) {
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

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <FormField
            control={form.control}
            name="password"
            rules={{
              required: 'Password is required',
              minLength: {
                value: 6,
                message: 'Password must be at least 6 characters',
              },
            }}
            render={({ field }) => (
              <FormItem>
                <FormLabel>New Password</FormLabel>
                <FormControl>
                  <PasswordInput
                    placeholder="Enter your new password"
                    autoComplete="new-password"
                    showStrength
                    disabled={isLoading || !token}
                    {...field}
                    value={password || ''}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="confirmPassword"
            rules={{
              required: 'Please confirm your password',
              validate: value =>
                value === password || 'Passwords do not match',
            }}
            render={({ field }) => (
              <FormItem>
                <FormLabel>Confirm New Password</FormLabel>
                <FormControl>
                  <PasswordInput
                    placeholder="Confirm your new password"
                    autoComplete="new-password"
                    disabled={isLoading || !token}
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
            disabled={isLoading || !token}
          >
            {isLoading ? 'Resetting password...' : 'Reset password'}
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