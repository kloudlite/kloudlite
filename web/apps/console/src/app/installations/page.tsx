import { redirect } from 'next/navigation'
import { getValidUserInstallations, type Installation } from '@/lib/console/supabase-storage-service'
import { getRegistrationSession } from '@/lib/console-auth'
import { InstallationsList } from '@/components/installations-list'
import { InstallationsHeader } from '@/components/installations-header'
import { GridContainer } from '@/components/grid-container'
import { PendingInvitationsBanner } from '@/components/pending-invitations-banner'

export default async function InstallationsPage() {
  const session = await getRegistrationSession()

  // Require authentication
  if (!session?.user) {
    redirect('/login')
  }

  // Fetch user's valid (non-expired) installations from database
  let installations: Installation[] = []
  try {
    installations = await getValidUserInstallations(session.user.id)
  } catch (error) {
    console.error('Error fetching installations:', error)
    installations = []
  }

  return (
    <div className="bg-background min-h-screen">
      <InstallationsHeader user={session.user} />
      <PendingInvitationsBanner />

      <main className="mx-auto max-w-7xl px-6 py-16">
        <GridContainer className="border-t">
          {/* Title Section */}
          <div className="border-b px-8 py-10">
            <h1 className="text-3xl font-bold tracking-tight">Installations</h1>
            <p className="text-muted-foreground mt-2 text-base">
              Manage your Kloudlite installations
            </p>
          </div>

          {/* Installations List with Filter */}
          <div className="px-8 py-10">
            <InstallationsList installations={installations} />
          </div>
        </GridContainer>
      </main>
    </div>
  )
}
