import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { SubscriptionManagement } from '@/components/billing/subscription-management'
import {
  getStripeCustomer,
  getSubscriptionItems,
  syncSubscriptionItemsFromStripe,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'
import { getTierConfig } from '@/lib/stripe-bootstrap'

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

  // TODO: detect currency from user locale or installation region
  const currency = 'usd'
  const tierConfig = await getTierConfig(currency)

  const customer = await getStripeCustomer(id)
  let items = await getSubscriptionItems(id)

  // Sync items from Stripe if DB is empty but subscription is active (webhook may not have fired)
  if (items.length === 0 && customer?.stripeSubscriptionId && customer.billingStatus === 'active') {
    await syncSubscriptionItemsFromStripe(id, customer.stripeSubscriptionId)
    items = await getSubscriptionItems(id)
  }

  return (
    <div className="space-y-6">
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <div className="mb-6">
          <h2 className="text-lg font-semibold text-foreground">Subscription</h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Manage your compute plan and billing
          </p>
        </div>
        <SubscriptionManagement
          installationId={id}
          customer={customer}
          items={items}
          tierConfig={tierConfig}
          currency={currency}
          isOwner={isOwner}
        />
      </div>
    </div>
  )
}
