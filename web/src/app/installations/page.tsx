import { redirect } from 'next/navigation'
import { getUserInstallations } from '@/lib/registration/supabase-storage-service'
import { getRegistrationSession } from '@/lib/registration-auth'
import { InstallationsList } from '@/components/installations-list'
import { InstallationsHeader } from '@/components/installations-header'

export default async function InstallationsPage() {
  const session = await getRegistrationSession()

  // Require authentication
  if (!session?.user) {
    redirect('/installations/login')
  }

  // Fetch user's installations from database
  let installations = []
  try {
    installations = await getUserInstallations(session.user.id)
  } catch (error) {
    console.error('Error fetching installations:', error)
    installations = []
  }

  return (
    <div className="min-h-screen bg-background">
      <InstallationsHeader user={session.user} />

      <main className="mx-auto max-w-7xl px-6 py-8">
        {/* Title Section */}
        <div className="mb-8">
          <div className="mb-6">
            <h1 className="text-2xl font-semibold">Installations</h1>
            <p className="text-muted-foreground mt-1.5 text-sm">
              Manage your Kloudlite installations
            </p>
          </div>

          {/* Installations List with Filter */}
          <InstallationsList installations={installations} />
        </div>
      </main>
    </div>
  )
}
