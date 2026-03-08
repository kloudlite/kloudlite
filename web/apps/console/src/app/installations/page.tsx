import { redirect } from 'next/navigation'
import {
  getValidOrgInstallations,
  getUserOrganizations,
  getBillingAccount,
  getUserPendingOrgInvitations,
  type Installation,
  type BillingAccount,
} from '@/lib/console/storage'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
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

  // Fetch pending invitations server-side
  const pendingInvitations = await getUserPendingOrgInvitations(session.user.email)

  // Get selected org (handles auto-creation for users with no orgs)
  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  const orgs = await getUserOrganizations(session.user.id)

  // Fetch installations for the selected org
  let installations: Installation[] = []
  let activeSubscriptions: Record<string, BillingAccount> = {}
  if (currentOrg) {
    try {
      installations = await getValidOrgInstallations(currentOrg.id)
      const billingAccount = await getBillingAccount(currentOrg.id)
      if (billingAccount) {
        installations.forEach((i) => {
          activeSubscriptions[i.id] = billingAccount
        })
      }
    } catch (error) {
      console.error('Error fetching installations:', error)
    }
  }

  return (
    <div className="bg-background h-screen flex flex-col">
      <InstallationsHeader
        user={session.user}
        orgs={orgs.map((o) => ({ id: o.id, name: o.name, slug: o.slug }))}
        currentOrgId={currentOrg?.id}
      />
      <PendingInvitationsBanner initialInvitations={pendingInvitations} />

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
