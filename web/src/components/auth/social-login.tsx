'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Github } from 'lucide-react'

interface SocialLoginProps {
  mode: 'signin' | 'signup'
}

export function SocialLogin({ mode }: SocialLoginProps) {
  const [isLoading, setIsLoading] = useState<string | null>(null)

  const handleSocialLogin = async (provider: string) => {
    setIsLoading(provider)
    // TODO: Implement actual OAuth flow
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsLoading(null)
  }

  const actionText = mode === 'signin' ? 'Sign in' : 'Sign up'

  return (
    <div className="space-y-3">
      <Button
        type="button"
        variant="outline"
        size="default"
        className="w-full h-11"
        onClick={() => handleSocialLogin('google')}
        disabled={isLoading !== null}
      >
        {isLoading === 'google' ? (
          'Loading...'
        ) : (
          <>
            <svg className="h-5 w-5 mr-2" viewBox="0 0 24 24">
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
            {actionText} with Google
          </>
        )}
      </Button>

      <Button
        type="button"
        variant="outline"
        size="default"
        className="w-full h-11"
        onClick={() => handleSocialLogin('microsoft')}
        disabled={isLoading !== null}
      >
        {isLoading === 'microsoft' ? (
          'Loading...'
        ) : (
          <>
            <svg className="h-5 w-5 mr-2" viewBox="0 0 23 23" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M11 11H0V0H11V11Z" fill="#F25022"/>
              <path d="M23 11H12V0H23V11Z" fill="#7FBA00"/>
              <path d="M11 23H0V12H11V23Z" fill="#00A4EF"/>
              <path d="M23 23H12V12H23V23Z" fill="#FFB900"/>
            </svg>
            {actionText} with Microsoft
          </>
        )}
      </Button>

      <Button
        type="button"
        variant="outline"
        size="default"
        className="w-full h-11"
        onClick={() => handleSocialLogin('github')}
        disabled={isLoading !== null}
      >
        {isLoading === 'github' ? (
          'Loading...'
        ) : (
          <>
            <Github className="h-5 w-5 mr-2" />
            {actionText} with GitHub
          </>
        )}
      </Button>
    </div>
  )
}