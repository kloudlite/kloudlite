import { redirect } from 'next/navigation'
import {
  getValidUserInstallations,
  getPendingInvoicesByInstallationIds,
  getActiveSubscriptionsByInstallationIds,
  type Installation,
  type Invoice,
  type Subscription,
} from '@/lib/console/storage'
import { getRegistrationSession } from '@/lib/console-auth'
import { InstallationsList } from '@/components/installations-list'
import { InstallationsHeader } from '@/components/installations-header'
import { PendingInvitationsBanner } from '@/components/pending-invitations-banner'

import { ScrollArea } from '@kloudlite/ui'

export default async function InstallationsPage() {
  const session = await getRegistrationSession()

  // Require authentication
  if (!session?.user) {
    redirect('/login')
  }

  // Fetch user's valid (non-expired) installations from database
  let installations: Installation[] = []
  let pendingInvoices: Record<string, Invoice> = {}
  let activeSubscriptions: Record<string, Subscription> = {}
  try {
    installations = await getValidUserInstallations(session.user.id)
    if (installations.length > 0) {
      const ids = installations.map((i) => i.id)
      ;[pendingInvoices, activeSubscriptions] = await Promise.all([
        getPendingInvoicesByInstallationIds(ids),
        getActiveSubscriptionsByInstallationIds(ids),
      ])
    }
  } catch (error) {
    console.error('Error fetching installations:', error)
    installations = []
  }

  return (
    <div className="bg-background h-screen flex flex-col">
      <InstallationsHeader user={session.user} />
      <PendingInvitationsBanner />

      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-12 py-8">
          <InstallationsList
            installations={installations}
            pendingInvoices={pendingInvoices}
            activeSubscriptions={activeSubscriptions}
          />
        </main>
      </ScrollArea>
    </div>
  )
}
