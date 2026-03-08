import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { CreditManagement } from '@/components/billing/credit-management'
import { cachedInstallationAccess } from '@/lib/console/cached-queries'

interface BillingPageProps {
  params: Promise<{ id: string }>
}

export default async function BillingPage({ params }: BillingPageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  let orgId: string
  let isOwner: boolean
  try {
    const auth = await cachedInstallationAccess(id)
    orgId = auth.orgId
    isOwner = auth.role === 'owner'
  } catch {
    redirect('/installations')
  }

  return (
    <div className="space-y-6">
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <div className="mb-6">
          <h2 className="text-lg font-semibold text-foreground">Billing</h2>
          <p className="text-muted-foreground mt-1 text-sm">
            Manage your credit balance and usage
          </p>
        </div>
        <CreditManagement orgId={orgId} isOwner={isOwner} />
      </div>
    </div>
  )
}
