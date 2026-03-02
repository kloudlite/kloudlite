import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { SubscriptionManagement } from '@/components/billing/subscription-management'
import { RazorpayProvider } from '@/components/razorpay-provider'
import {
  getPlans,
  getSubscriptionsByInstallation,
  getInvoicesByInstallation,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'

interface SubscriptionPageProps {
  params: Promise<{ id: string }>
}

export default async function SubscriptionPage({ params }: SubscriptionPageProps) {
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

  const isOwner = role === 'owner'

  const [plans, subscriptions, invoices] = await Promise.all([
    getPlans(),
    getSubscriptionsByInstallation(id),
    getInvoicesByInstallation(id),
  ])

  return (
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
  )
}
