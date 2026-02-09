import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getValidUserInstallations, type Installation } from '@/lib/console/storage'
import { getRegistrationSession } from '@/lib/console-auth'
import { InstallationsList } from '@/components/installations-list'
import { InstallationsHeader } from '@/components/installations-header'
import { PendingInvitationsBanner } from '@/components/pending-invitations-banner'
import { Button, ScrollArea } from '@kloudlite/ui'
import { Plus } from 'lucide-react'

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
    <div className="bg-background h-screen flex flex-col">
      <InstallationsHeader user={session.user} />
      <PendingInvitationsBanner />

      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-12 py-8">
          {/* Page Header */}
          <div className="flex items-center justify-between border-b border-foreground/10 pb-6 mb-6">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight text-foreground">Installations</h1>
              <p className="text-muted-foreground mt-1 text-sm">
                Manage and monitor your cloud deployments
              </p>
            </div>
            <Link href="/installations/new-kl-cloud">
              <Button size="default">
                <Plus className="h-4 w-4" />
                New Installation
              </Button>
            </Link>
          </div>

          {/* Installations List with Filter */}
          <InstallationsList installations={installations} />
        </main>
      </ScrollArea>
    </div>
  )
}
