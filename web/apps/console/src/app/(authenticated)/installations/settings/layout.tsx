import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { NavTabs } from '@/components/nav-tabs'

interface LayoutProps {
  children: React.ReactNode
}

export default async function SettingsLayout({ children }: LayoutProps) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)

  return (
    <main className="mx-auto max-w-6xl px-6 lg:px-12 py-10">
      {/* Title Section */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-foreground">
          {currentOrg?.name ? `${currentOrg.name}'s Settings` : 'Settings'}
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Manage your organization and billing
        </p>
      </div>

      {/* Tabs Navigation */}
      <NavTabs tabs={[
        { id: 'organization', label: 'Organization', href: '/installations/settings/organization' },
        { id: 'billing', label: 'Billing', href: '/installations/settings/billing' },
      ]} />

      {/* Content */}
      <div className="mt-6">{children}</div>
    </main>
  )
}
