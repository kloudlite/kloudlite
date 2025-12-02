import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { ArtifactsNav } from './_components/artifacts-nav'

interface LayoutProps {
  children: React.ReactNode
}

export default async function ArtifactsLayout({ children }: LayoutProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Title Section */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold">Artifacts</h1>
        <p className="text-muted-foreground mt-1.5 text-sm">
          Browse and manage container images pushed from your workspaces
        </p>
      </div>

      {/* Navigation Tabs */}
      <ArtifactsNav />

      {/* Page Content */}
      <div className="mt-6">{children}</div>
    </main>
  )
}
