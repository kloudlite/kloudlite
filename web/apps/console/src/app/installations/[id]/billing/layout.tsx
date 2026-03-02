import { ReactNode } from 'react'
import { BillingSidebar } from '@/components/billing-sub-tabs'

interface BillingLayoutProps {
  children: ReactNode
  params: Promise<{ id: string }>
}

export default async function BillingLayout({ children, params }: BillingLayoutProps) {
  const { id } = await params

  return (
    <div className="border border-foreground/10 rounded-lg p-6 bg-background">
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-foreground">Billing & Compute</h2>
        <p className="text-muted-foreground mt-1 text-sm">
          Manage your subscriptions and payment methods
        </p>
      </div>

      <div className="flex gap-8">
        <aside className="w-44 shrink-0 border-r border-foreground/10 pr-6">
          <BillingSidebar installationId={id} />
        </aside>
        <div className="flex-1 min-w-0">{children}</div>
      </div>
    </div>
  )
}
