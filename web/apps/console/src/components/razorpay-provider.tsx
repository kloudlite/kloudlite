'use client'

import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react'

declare global {
  interface Window {
    Razorpay?: new (options: RazorpayCheckoutOptions) => {
      open(): void
      on(event: string, handler: (resp: Record<string, unknown>) => void): void
    }
  }
}

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

export type { RazorpayCheckoutOptions }

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
    if (window.Razorpay) {
      setIsLoaded(true)
      return
    }

    const script = document.createElement('script')
    script.src = CHECKOUT_SRC
    script.async = true
    script.onload = () => {
      console.log('[Razorpay] Checkout script loaded successfully')
      setIsLoaded(true)
    }
    script.onerror = () => console.error('[Razorpay] Failed to load checkout script')
    document.head.appendChild(script)
  }, [])

  const openCheckout = useCallback(
    (options: RazorpayCheckoutOptions) => {
      const RazorpayClass = window.Razorpay
      if (!RazorpayClass) {
        throw new Error('Razorpay checkout not loaded')
      }

      console.log('[Razorpay] Opening checkout with:', {
        key: options.key ? `${options.key.slice(0, 12)}...` : 'MISSING',
        order_id: options.order_id,
        amount: options.amount,
        currency: options.currency,
        hasHandler: !!options.handler,
        hasOndismiss: !!options.modal?.ondismiss,
      })

      const rzp = new RazorpayClass(options)

      rzp.on('payment.failed', (resp: Record<string, unknown>) => {
        console.error('[Razorpay] payment.failed event:', resp?.error)
      })

      rzp.open()
      console.log('[Razorpay] rzp.open() called')
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
