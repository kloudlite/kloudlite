import Script from 'next/script'
import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { BillingContent } from '@/components/billing-content'
import {
  getPlans,
  getSubscriptionByInstallation,
  getInvoicesByInstallation,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'

interface BillingPageProps {
  searchParams: Promise<{ installation?: string }>
}

export default async function BillingPage({ searchParams }: BillingPageProps) {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  const params = await searchParams
  const installationId = params.installation

  if (!installationId) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Select an installation to manage billing.</p>
      </div>
    )
  }

  const installation = await getInstallationById(installationId)
  if (!installation) {
    redirect('/installations')
  }

  // Check ownership
  let role = await getMemberRole(installationId, session.user.id)
  if (!role && installation.userId === session.user.id) {
    role = 'owner'
  }

  const isOwner = role === 'owner'

  const [plans, subscription, invoices] = await Promise.all([
    getPlans(),
    getSubscriptionByInstallation(installationId),
    getInvoicesByInstallation(installationId),
  ])

  return (
    <>
      <Script src="https://checkout.razorpay.com/v1/checkout.js" strategy="lazyOnload" />
      <BillingContent
        installationId={installationId}
        plans={plans}
        subscription={subscription}
        invoices={invoices}
        isOwner={isOwner}
        userEmail={session.user.email}
        userName={session.user.name}
      />
    </>
  )
}
