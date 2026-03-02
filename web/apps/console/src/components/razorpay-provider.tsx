'use client'

import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react'

interface RazorpayCheckoutOptions {
  key?: string
  subscription_id?: string
  order_id?: string
  amount?: number
  currency?: string
  name?: string
  description?: string
  prefill?: { name?: string; email?: string; contact?: string }
  notes?: Record<string, string>
  theme?: { color?: string }
  handler?: (response: Record<string, string>) => void
  modal?: { ondismiss?: () => void }
}

interface RazorpayContextValue {
  isLoaded: boolean
  openCheckout: (options: RazorpayCheckoutOptions) => void
}

const RazorpayContext = createContext<RazorpayContextValue | null>(null)

const CHECKOUT_SRC = 'https://checkout.razorpay.com/v1/checkout.js'

export function RazorpayProvider({ children }: { children: React.ReactNode }) {
  const [isLoaded, setIsLoaded] = useState(false)
  const scriptRef = useRef(false)

  useEffect(() => {
    if (scriptRef.current) return
    scriptRef.current = true

    // Already loaded (e.g. HMR)
    if ((window as any).Razorpay) {
      setIsLoaded(true)
      return
    }

    const script = document.createElement('script')
    script.src = CHECKOUT_SRC
    script.async = true
    script.onload = () => setIsLoaded(true)
    script.onerror = () => console.error('Failed to load Razorpay checkout script')
    document.head.appendChild(script)
  }, [])

  const openCheckout = useCallback(
    (options: RazorpayCheckoutOptions) => {
      if (!(window as any).Razorpay) {
        throw new Error('Razorpay checkout not loaded')
      }
      const rzp = new (window as any).Razorpay(options)
      rzp.open()
    },
    [],
  )

  return (
    <RazorpayContext.Provider value={{ isLoaded, openCheckout }}>
      {children}
    </RazorpayContext.Provider>
  )
}

export function useRazorpay() {
  const ctx = useContext(RazorpayContext)
  if (!ctx) {
    throw new Error('useRazorpay must be used within <RazorpayProvider>')
  }
  return ctx
}
