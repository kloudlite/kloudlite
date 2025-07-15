'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { Link } from '@/components/ui/link'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
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
import { LoginCredentials } from '@/lib/auth/types'
import { loginAction } from '@/actions/auth/login'
import { AlertCircle, CheckCircle2 } from 'lucide-react'
import { SocialLogin } from './social-login'
import { AuthDivider } from './auth-divider'
import { SSOLogin } from './sso-login'

export function LoginForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const form = useForm<LoginCredentials>({
    defaultValues: {
      email: '',
      password: '',
      rememberMe: false,
    },
  })

  useEffect(() => {
    if (searchParams.get('reset') === 'success') {
      setSuccess('Your password has been reset successfully. Please sign in with your new password.')
    } else if (searchParams.get('verified') === 'success') {
      setSuccess('Your email has been verified successfully. You can now sign in.')
    }
  }, [searchParams])

  const onSubmit = async (data: LoginCredentials) => {
    setIsLoading(true)
    setError(null)

    try {
      const result = await loginAction(data)
      
      if (result.success) {
        router.push('/dashboard')
        router.refresh()
      } else {
        setError(result.error || 'An error occurred during login')
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

      <SocialLogin mode="signin" />
      
      <SSOLogin />
      
      <AuthDivider />

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
                  disabled={isLoading}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

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
              <div className="flex items-center justify-between">
                <FormLabel>Password</FormLabel>
                <Link href="/auth/forgot-password" size="sm">
                  Forgot password?
                </Link>
              </div>
              <FormControl>
                <PasswordInput
                  placeholder="Enter your password"
                  autoComplete="current-password"
                  disabled={isLoading}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

          <FormField
          control={form.control}
          name="rememberMe"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center space-x-3 space-y-0">
              <FormControl>
                <Checkbox
                  checked={field.value}
                  onCheckedChange={field.onChange}
                  disabled={isLoading}
                />
              </FormControl>
              <FormLabel className="font-normal cursor-pointer select-none">
                Remember me for 30 days
              </FormLabel>
            </FormItem>
          )}
          />

          <Button type="submit" className="w-full" size="auth" disabled={isLoading}>
          {isLoading ? 'Signing in...' : 'Sign in'}
        </Button>

          <div className="text-center text-sm space-y-1">
            <span className="text-muted-foreground">Don't have an account?</span>{' '}
            <Link href="/auth/signup">
              Sign up
            </Link>
          </div>
        </form>
      </Form>
    </div>
  )
}