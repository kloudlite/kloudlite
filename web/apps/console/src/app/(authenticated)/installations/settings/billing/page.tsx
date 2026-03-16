import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getOrgMemberRole } from '@/lib/console/storage'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { CreditManagement } from '@/components/billing/credit-management'

export default async function BillingSettingsPage() {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  if (!currentOrg) redirect('/installations')

  const userRole = await getOrgMemberRole(currentOrg.id, session.user.id)
  if (!userRole) redirect('/installations')

  const isOwner = userRole === 'owner'

  return <CreditManagement orgId={currentOrg.id} isOwner={isOwner} />
}
