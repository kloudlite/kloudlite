import { redirect } from 'next/navigation'
import {
  getValidUserInstallations,
  getActiveSubscriptionsByInstallationIds,
  type Installation,
  type StripeCustomer,
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
  let activeSubscriptions: Record<string, StripeCustomer> = {}
  try {
    installations = await getValidUserInstallations(session.user.id)
    if (installations.length > 0) {
      const ids = installations.map((i) => i.id)
      activeSubscriptions = await getActiveSubscriptionsByInstallationIds(ids)
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
            activeSubscriptions={activeSubscriptions}
          />
        </main>
      </ScrollArea>
    </div>
  )
}
