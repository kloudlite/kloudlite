'use client'

import { AlertTriangle, Loader2 } from 'lucide-react'
import { formatCurrency } from '@/lib/billing-utils'
import type { Invoice } from '@/lib/console/storage'

interface PaymentDueBannerProps {
  pendingInvoice: Invoice
  paying: boolean
  onPayNow: () => void
}

export function PaymentDueBanner({ pendingInvoice, paying, onPayNow }: PaymentDueBannerProps) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border border-amber-500/20 bg-amber-500/10 p-4 mb-6">
      <div className="flex items-center gap-3">
        <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0" />
        <div>
          <p className="text-sm font-medium text-amber-800 dark:text-amber-200">Payment Due</p>
          <p className="text-xs text-amber-700 dark:text-amber-300 mt-0.5">
            {formatCurrency(pendingInvoice.amount, pendingInvoice.currency)} due for your subscription renewal
          </p>
        </div>
      </div>
      <button
        onClick={onPayNow}
        disabled={paying}
        className="shrink-0 rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white hover:bg-amber-700 transition-colors disabled:opacity-50"
      >
        {paying ? (
          <span className="flex items-center gap-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            Processing...
          </span>
        ) : (
          'Pay Now'
        )}
      </button>
    </div>
  )
}
