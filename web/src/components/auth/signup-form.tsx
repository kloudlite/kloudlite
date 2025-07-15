'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link } from '@/components/ui/link'
import { SignupCredentials } from '@/lib/auth/types'
import { signupAction } from '@/actions/auth/signup'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { PasswordInput } from '@/components/auth/password-input'
import { AlertCircle } from 'lucide-react'
import { SocialLogin } from './social-login'
import { AuthDivider } from './auth-divider'
import { SSOLogin } from './sso-login'

export function SignupForm() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<SignupCredentials>({
    defaultValues: {
      name: '',
      email: '',
      password: '',
      acceptTerms: false
    }
  })

  const password = form.watch('password')

  const onSubmit = async (data: SignupCredentials) => {
    setIsLoading(true)
    setError(null)

    const result = await signupAction(data)

    if (!result.success) {
      setError(result.error || 'An error occurred during signup')
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

      <SocialLogin mode="signup" />
      
      <SSOLogin />
      
      <AuthDivider />

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">

          <FormField
          control={form.control}
          name="name"
          rules={{
            required: 'Name is required',
            minLength: {
              value: 2,
              message: 'Name must be at least 2 characters'
            }
          }}
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input
                  type="text"
                  placeholder="John Doe"
                  autoComplete="name"
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
          name="email"
          rules={{
            required: 'Email is required',
            pattern: {
              value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
              message: 'Invalid email address'
            }
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
              message: 'Password must be at least 6 characters'
            }
          }}
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder="Create a password"
                  autoComplete="new-password"
                  showStrength
                  disabled={isLoading}
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
          name="acceptTerms"
          rules={{
            required: 'You must accept the terms and conditions'
          }}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center space-x-3 space-y-0">
              <FormControl>
                <Checkbox
                  checked={field.value}
                  onCheckedChange={field.onChange}
                  disabled={isLoading}
                />
              </FormControl>
              <div className="space-y-1 leading-none">
                <FormLabel className="font-normal cursor-pointer select-none">
                  I accept the{' '}
                  <Link href="/terms" size="sm">
                    terms and conditions
                  </Link>
                </FormLabel>
                <FormMessage />
              </div>
            </FormItem>
          )}
          />

          <Button type="submit" className="w-full" size="auth" disabled={isLoading}>
          {isLoading ? 'Creating account...' : 'Create account'}
        </Button>

          <div className="text-center text-sm space-y-1">
            <span className="text-muted-foreground">Already have an account?</span>{' '}
            <Link href="/auth/login">
              Sign in
            </Link>
          </div>
        </form>
      </Form>
    </div>
  )
}