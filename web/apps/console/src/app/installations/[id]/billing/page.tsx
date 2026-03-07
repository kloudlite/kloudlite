import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { SubscriptionManagement } from '@/components/billing/subscription-management'
import {
  getStripeCustomer,
  getSubscriptionItems,
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

  const [customer, items] = await Promise.all([
    getStripeCustomer(id),
    getSubscriptionItems(id),
  ])

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
          isOwner={isOwner}
        />
      </div>
    </div>
  )
}
