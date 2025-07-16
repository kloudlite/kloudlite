'use client'

import { useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useSession } from 'next-auth/react'

export default function AuthCallbackPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { data: session, status } = useSession()
  
  useEffect(() => {
    // Get the intended callback URL
    const callbackUrl = searchParams.get('callbackUrl') || '/teams'
    
    // If we have a session, redirect to the callback URL
    if (status === 'authenticated' && session) {
      console.log('Session authenticated, redirecting to:', callbackUrl)
      router.push(callbackUrl)
    } else if (status === 'unauthenticated') {
      console.log('Authentication failed, redirecting to login')
      router.push('/auth/login')
    }
  }, [status, session, router, searchParams])

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p className="mt-4 text-sm text-muted-foreground">
          Completing sign in...
        </p>
      </div>
    </div>
  )
}