import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { getUserOrganizations } from '@/lib/console/storage'
import { InstallationsHeader } from '@/components/installations-header'
import { InstallationSettingsTabs } from '@/components/installation-settings-tabs'
import { ScrollArea } from '@kloudlite/ui'

interface LayoutProps {
  children: React.ReactNode
}

export default async function SettingsLayout({ children }: LayoutProps) {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  const orgs = await getUserOrganizations(session.user.id)

  return (
    <div className="bg-background flex h-screen flex-col">
      <InstallationsHeader
        user={session.user}
        orgs={orgs.map((o) => ({ id: o.id, name: o.name, slug: o.slug }))}
        currentOrgId={currentOrg?.id}
      />

      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-12 py-10">
          {/* Title Section */}
          <div className="mb-6">
            <h1 className="text-4xl lg:text-5xl font-bold tracking-tight text-foreground leading-[1.1]">Settings</h1>
            <p className="text-muted-foreground mt-2 text-[1.0625rem] leading-relaxed">
              Manage your organization and billing
            </p>
          </div>

          {/* Tabs Navigation */}
          <InstallationSettingsTabs />

          {/* Content */}
          <div className="mt-6">{children}</div>
        </main>
      </ScrollArea>
    </div>
  )
}
