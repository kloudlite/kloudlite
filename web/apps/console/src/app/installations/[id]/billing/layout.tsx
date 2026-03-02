import { ReactNode } from 'react'
import { BillingSidebar } from '@/components/billing-sub-tabs'

interface BillingLayoutProps {
  children: ReactNode
  params: Promise<{ id: string }>
}

export default async function BillingLayout({ children, params }: BillingLayoutProps) {
  const { id } = await params

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold">Billing & Compute</h2>
        <p className="text-muted-foreground mt-1 text-base">
          Manage your subscriptions and payment methods
        </p>
      </div>

      <div className="flex gap-8">
        <aside className="w-48 shrink-0">
          <BillingSidebar installationId={id} />
        </aside>
        <div className="flex-1 min-w-0">{children}</div>
      </div>
    </div>
  )
}
