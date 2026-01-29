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

      <main className="mx-auto max-w-7xl px-6 lg:px-12 py-16">
        {/* Title Section */}
        <div className="mb-10">
          <h1 className="text-4xl lg:text-5xl font-bold tracking-tight text-foreground leading-[1.1]">Installations</h1>
          <p className="text-muted-foreground mt-3 text-[1.0625rem] leading-relaxed">
            Manage and monitor your cloud deployments
          </p>
        </div>

        {/* Installations List with Filter */}
        <InstallationsList installations={installations} />
      </main>
    </div>
  )
}
