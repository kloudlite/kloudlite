import { ReactNode } from 'react'
import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { getUserOrganizations } from '@/lib/console/storage'
import { InstallationsHeader } from '@/components/installations-header'
import { InstallationLayoutShell } from '@/components/installation-layout'

export default async function NewByocLayout({ children }: { children: ReactNode }) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  const orgs = await getUserOrganizations(session.user.id)

  return (
    <div className="bg-background h-screen flex flex-col">
      <InstallationsHeader
        user={session.user}
        orgs={orgs.map((o) => ({ id: o.id, name: o.name, slug: o.slug }))}
        currentOrgId={currentOrg?.id}
      />
      <InstallationLayoutShell>{children}</InstallationLayoutShell>
    </div>
  )
}
