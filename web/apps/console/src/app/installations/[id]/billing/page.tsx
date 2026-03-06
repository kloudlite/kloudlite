import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { SubscriptionManagement } from '@/components/billing/subscription-management'
import { InvoiceHistory } from '@/components/billing/invoice-history'
import { RazorpayProvider } from '@/components/razorpay-provider'
import {
  getPlans,
  getSubscriptionsByInstallation,
  getInvoicesByInstallation,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'

interface BillingPageProps {
  params: Promise<{ id: string }>
}

export default async function BillingPage({ params }: BillingPageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  const installation = await getInstallationById(id)
  if (!installation) {
    redirect('/installations')
  }

  let role = await getMemberRole(id, session.user.id)
  if (!role && installation.userId === session.user.id) {
    role = 'owner'
  }
  if (!role) {
    redirect('/installations')
  }

  const isOwner = role === 'owner'

  const [plans, subscriptions, invoices] = await Promise.all([
    getPlans(),
    getSubscriptionsByInstallation(id),
    getInvoicesByInstallation(id),
  ])

  return (
    <div className="space-y-6">
      {/* Subscription Card */}
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <div className="mb-6">
          <h2 className="text-lg font-semibold text-foreground">Subscription</h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Manage your compute plan and billing
          </p>
        </div>
        <RazorpayProvider>
          <SubscriptionManagement
            installationId={id}
            plans={plans}
            subscriptions={subscriptions}
            invoices={invoices}
            isOwner={isOwner}
            userEmail={session.user.email}
            userName={session.user.name}
          />
        </RazorpayProvider>
      </div>

      {/* Invoice History Card */}
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <div className={invoices.length > 0 ? 'mb-6' : ''}>
          <h2 className="text-lg font-semibold text-foreground">Invoice History</h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Past payments and billing records
          </p>
        </div>
        {invoices.length > 0 ? (
          <InvoiceHistory invoices={invoices} />
        ) : (
          <p className="text-muted-foreground text-sm mt-4">No invoices yet.</p>
        )}
      </div>
    </div>
  )
}
