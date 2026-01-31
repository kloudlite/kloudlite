'use client'

/**
 * Magic Link Login Form with Cloudflare Turnstile Captcha (Invisible Mode)
 */

import { useState, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@kloudlite/ui'
import Turnstile from 'react-turnstile'

const formSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
})

type FormData = z.infer<typeof formSchema>

interface MagicLinkFormProps {
  siteKey: string
  onSuccess?: () => void
}

export function MagicLinkForm({ siteKey, onSuccess }: MagicLinkFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [captchaToken, setCaptchaToken] = useState<string | null>(null)
  const turnstileRef = useRef<any>(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
    getValues,
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
  })

  if (!siteKey) {
    return (
      <div className="border-destructive bg-destructive/10 text-destructive rounded-lg border p-4 text-sm">
        Captcha is not configured. Please contact support.
      </div>
    )
  }

  const onSubmit = async (data: FormData) => {
    const isDevelopment = process.env.NODE_ENV === 'development'

    // In production, require captcha token
    if (!isDevelopment && !captchaToken) {
      setError('Please wait for security verification to complete.')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/auth/magic-link/send', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: data.email,
          captchaToken: captchaToken || 'dev-bypass',
        }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Failed to send magic link')
      }

      setSuccess(true)
      onSuccess?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
      // Reset captcha on error
      setCaptchaToken(null)
      if (turnstileRef.current) {
        turnstileRef.current.reset()
      }
    } finally {
      setIsLoading(false)
    }
  }

  if (success) {
    return (
      <div className="space-y-4">
        <div className="border-success/30 bg-success/5 rounded-lg border p-4">
          <div className="flex items-start gap-3">
            <svg
              className="h-5 w-5 flex-shrink-0 mt-0.5 text-success"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 19v-8.93a2 2 0 01.89-1.664l7-4.666a2 2 0 012.22 0l7 4.666A2 2 0 0121 10.07V19M3 19a2 2 0 002 2h14a2 2 0 002-2M3 19l6.75-4.5M21 19l-6.75-4.5M3 10l6.75 4.5M21 10l-6.75 4.5m0 0l-1.14.76a2 2 0 01-2.22 0l-1.14-.76"
              />
            </svg>
            <div className="flex-1 text-foreground">
              <h3 className="font-semibold">Check your email!</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                We sent a magic link to <strong className="text-foreground">{getValues('email')}</strong>
              </p>
              <p className="mt-2 text-xs text-muted-foreground">
                The link will expire in 15 minutes.
              </p>
            </div>
          </div>
        </div>

        <button
          type="button"
          onClick={() => {
            setSuccess(false)
            setCaptchaToken(null)
            if (turnstileRef.current) {
              turnstileRef.current.reset()
            }
          }}
          className="text-sm text-muted-foreground hover:text-foreground underline transition-colors"
        >
          Send another link
        </button>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <label
          htmlFor="email"
          className="block text-sm font-medium text-foreground mb-2"
        >
          Email address
        </label>
        <input
          {...register('email')}
          type="email"
          id="email"
          autoComplete="email"
          disabled={isLoading}
          className="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-foreground/50 focus:outline-none focus:ring-2 focus:ring-foreground/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          placeholder="you@example.com"
        />
        {errors.email && (
          <p className="mt-1.5 text-xs text-destructive">{errors.email.message}</p>
        )}
      </div>

      {/* Invisible Turnstile widget */}
      {process.env.NODE_ENV !== 'development' && (
        <Turnstile
          ref={turnstileRef}
          sitekey={siteKey}
          onVerify={(token) => {
            setCaptchaToken(token)
          }}
          onError={() => {
            setError('Captcha verification failed. Please try again.')
          }}
          onExpire={() => {
            setCaptchaToken(null)
          }}
          size="invisible"
        />
      )}

      {error && (
        <div className="border-destructive bg-destructive/10 text-destructive border p-3 text-sm rounded-lg">
          {error}
        </div>
      )}

      <Button
        type="submit"
        disabled={isLoading}
        size="lg"
        className="w-full gap-3 text-base font-medium bg-primary text-primary-foreground hover:bg-primary/90 border-primary transition-all duration-200 shadow-sm hover:shadow-md group"
      >
        {isLoading ? (
          <>
            <svg
              className="animate-spin h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            <span className="flex-1 text-left">Sending magic link...</span>
          </>
        ) : (
          <>
            <svg
              className="h-5 w-5 transition-transform group-hover:scale-110"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 19v-8.93a2 2 0 01.89-1.664l7-4.666a2 2 0 012.22 0l7 4.666A2 2 0 0121 10.07V19M3 19a2 2 0 002 2h14a2 2 0 002-2M3 19l6.75-4.5M21 19l-6.75-4.5M3 10l6.75 4.5M21 10l-6.75 4.5m0 0l-1.14.76a2 2 0 01-2.22 0l-1.14-.76"
              />
            </svg>
            <span className="flex-1 text-left">Send magic link</span>
            <svg
              className="h-4 w-4 opacity-80 group-hover:opacity-100 group-hover:translate-x-0.5 transition-all"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </>
        )}
      </Button>

      <div className="space-y-2">
        <p className="text-xs text-center text-muted-foreground">
          We'll send you a link to sign in without a password
        </p>
        {process.env.NODE_ENV === 'development' && (
          <p className="text-xs text-center text-warning/80">
            [Dev Mode: Captcha bypassed]
          </p>
        )}
      </div>
    </form>
  )
}
