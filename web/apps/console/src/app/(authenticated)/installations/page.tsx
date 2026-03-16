import {
  getValidOrgInstallations,
  getBillingAccount,
  getUserPendingOrgInvitations,
  type Installation,
  type BillingAccount,
} from '@/lib/console/storage'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { InstallationsList } from '@/components/installations-list'
import { PendingInvitationsBanner } from '@/components/pending-invitations-banner'

export default async function InstallationsPage() {
  const session = await getRegistrationSession()
  if (!session?.user) return null

  const pendingInvitations = await getUserPendingOrgInvitations(session.user.email)
  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)

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
    <>
      <PendingInvitationsBanner initialInvitations={pendingInvitations} />
      <main className="mx-auto max-w-6xl px-6 lg:px-12 py-10">
        <InstallationsList
          installations={installations}
          activeSubscriptions={activeSubscriptions}
        />
      </main>
    </>
  )
}
