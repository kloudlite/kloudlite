import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getBillingAccount,
  getSubscriptionItems,
  getOrgMemberRole,
  getValidOrgInstallations,
  syncSubscriptionItemsFromStripe,
  upsertBillingAccount,
} from '@/lib/console/storage'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { getTierConfig } from '@/lib/stripe-bootstrap'
import { SubscriptionManagement } from '@/components/billing/subscription-management'
import { CreditCard, AlertTriangle } from 'lucide-react'

/**
 * When the billing account exists but is 'incomplete', check Stripe directly
 * for an active subscription. This handles the webhook race condition where
 * the redirect from Stripe checkout lands before the webhook fires, and also
 * handles local dev where STRIPE_WEBHOOK_SECRET is not configured.
 */
async function syncBillingFromStripeIfNeeded(
  orgId: string,
  billingAccount: Awaited<ReturnType<typeof getBillingAccount>>,
): Promise<Awaited<ReturnType<typeof getBillingAccount>>> {
  if (!billingAccount?.stripeCustomerId) return billingAccount
  if (billingAccount.billingStatus === 'active' || billingAccount.billingStatus === 'cancelled') {
    return billingAccount
  }

  try {
    const { getStripe } = await import('@/lib/stripe')
    const stripe = getStripe()

    // List active subscriptions for this customer (no deep expand on list)
    const subscriptions = await stripe.subscriptions.list({
      customer: billingAccount.stripeCustomerId,
      status: 'active',
      limit: 1,
    })

    if (subscriptions.data.length === 0) return billingAccount

    // Retrieve subscription with expanded product data
    const subscription = await stripe.subscriptions.retrieve(subscriptions.data[0].id, {
      expand: ['items.data.price.product'],
    })
    const firstItem = subscription.items.data[0]
    const periodEnd = firstItem?.current_period_end ?? null

    // Update the billing account in DB
    await upsertBillingAccount({
      orgId,
      stripeCustomerId: billingAccount.stripeCustomerId,
      stripeSubscriptionId: subscription.id,
      billingStatus: subscription.cancel_at_period_end ? 'cancelled' : 'active',
      currentPeriodEnd: periodEnd ? new Date(periodEnd * 1000).toISOString() : null,
    })

    // Sync subscription items from Stripe
    await syncSubscriptionItemsFromStripe(orgId, subscription.id)

    // Re-fetch the updated billing account
    return await getBillingAccount(orgId)
  } catch (err) {
    console.error('[billing-page] Failed to sync from Stripe:', err)
    return billingAccount
  }
}

interface BillingSettingsPageProps {
  searchParams: Promise<{ checkout?: string }>
}

export default async function BillingSettingsPage({ searchParams }: BillingSettingsPageProps) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  if (!currentOrg) redirect('/installations')
  const userRole = await getOrgMemberRole(currentOrg.id, session.user.id)
  if (!userRole) redirect('/installations')

  const params = await searchParams
  const isOwner = userRole === 'owner'
  let billingAccount = await getBillingAccount(currentOrg.id)

  // Handle webhook race condition: if billing account is incomplete, check Stripe directly
  billingAccount = await syncBillingFromStripeIfNeeded(currentOrg.id, billingAccount)

  const installations = await getValidOrgInstallations(currentOrg.id)
  const hasInstallations = installations.length > 0
  const hasSubscription = billingAccount?.billingStatus === 'active' || billingAccount?.billingStatus === 'cancelled'

  // Fetch subscription details if there's an active subscription
  let subscriptionItems = billingAccount?.stripeSubscriptionId
    ? await getSubscriptionItems(currentOrg.id)
    : []

  // If items are empty but we have a subscription, sync items from Stripe
  if (subscriptionItems.length === 0 && billingAccount?.stripeSubscriptionId) {
    await syncSubscriptionItemsFromStripe(currentOrg.id, billingAccount.stripeSubscriptionId)
    subscriptionItems = await getSubscriptionItems(currentOrg.id)
  }

  const tierConfig = await getTierConfig()

  return (
    <div className="space-y-6">
      <div className="mb-5">
        <h2 className="text-xl font-semibold">Billing &amp; Subscription</h2>
        <p className="text-muted-foreground mt-1 text-base">
          Manage your organization&apos;s subscription and billing
        </p>
      </div>

      {/* Checkout cancelled banner */}
      {params.checkout === 'cancelled' && (
        <div className="flex items-center gap-3 rounded-lg border border-yellow-500/30 bg-yellow-500/5 px-4 py-3 text-sm">
          <AlertTriangle className="size-4 text-yellow-600 dark:text-yellow-400 shrink-0" />
          <p className="text-foreground">
            Checkout was cancelled. You can subscribe again when you&apos;re ready.
          </p>
        </div>
      )}

      {/* No installations and no existing subscription — show empty state */}
      {!hasInstallations && !hasSubscription ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted/50 p-4 mb-4">
            <CreditCard className="h-8 w-8 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-semibold mb-2">No billing required yet</h3>
          <p className="text-muted-foreground text-sm max-w-md mb-6">
            Billing is set up when you create your first installation. Create an installation to get started.
          </p>
          <Link
            href="/installations"
            className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Go to Installations
          </Link>
        </div>
      ) : (
        <SubscriptionManagement
          orgId={currentOrg.id}
          customer={billingAccount}
          items={subscriptionItems}
          tierConfig={tierConfig}
          currency="usd"
          isOwner={isOwner}
        />
      )}
    </div>
  )
}
