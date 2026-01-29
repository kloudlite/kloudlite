import { redirect } from 'next/navigation'
import { ReactNode } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById, getMemberRole } from '@/lib/console/supabase-storage-service'
import { InstallationsHeader } from '@/components/installations-header'
import { InstallationDetailsTabs } from '@/components/installation-details-tabs'
import { ScrollArea } from '@kloudlite/ui'

interface LayoutProps {
  children: ReactNode
  params: Promise<{ id: string }>
}

export default async function InstallationLayout({ children, params }: LayoutProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  const installation = await getInstallationById(id)

  if (!installation) {
    redirect('/installations')
  }

  // Check if user has access to this installation
  const userRole = await getMemberRole(id, session.user.id)

  if (!userRole) {
    redirect('/installations')
  }

  const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'
  const installationDomain = installation.subdomain ? `${installation.subdomain}.${domain}` : undefined

  return (
    <div className="bg-background h-screen flex flex-col">
      <InstallationsHeader
        user={session.user}
        installationName={installation.name}
        installationDomain={installationDomain}
      />

      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-8 py-8">
          {/* Back Button */}
          <div className="mb-8">
            <Link
              href="/installations"
              className="group inline-flex items-center gap-2 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm"
            >
              <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
              <span className="relative">
                Back to Installations
                <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
              </span>
            </Link>
          </div>

          {/* Page Header */}
          <div className="border-b border-foreground/10 pb-6 mb-8">
            {/* Installation Name and Domain */}
            <div className="flex items-start justify-between mb-4">
              <div className="flex-1">
                <h1 className="text-2xl font-semibold tracking-tight text-foreground mb-2">
                  Installation: {installation.name}
                </h1>
                {installation.description && (
                  <p className="text-muted-foreground text-sm leading-relaxed">
                    {installation.description}
                  </p>
                )}
              </div>
              {installationDomain && (
                <div className="ml-6 flex-shrink-0">
                  <a
                    href={`https://${installationDomain}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-2 px-3 py-1.5 text-xs font-mono text-primary bg-primary/5 border border-primary/20 rounded-md hover:bg-primary/10 transition-colors"
                  >
                    {installationDomain}
                    <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  </a>
                </div>
              )}
            </div>

            {/* Installation Metadata */}
            <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <span>Created {new Date(installation.createdAt).toLocaleDateString()}</span>
              </div>
              {installation.lastHealthCheck && (
                <div className="flex items-center gap-1.5">
                  <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>Last checked {new Date(installation.lastHealthCheck).toLocaleDateString()}</span>
                </div>
              )}
            </div>
          </div>

          {/* Tabs */}
          <div className="mb-8">
            <InstallationDetailsTabs installationId={id} />
          </div>

          {/* Page Content */}
          {children}
        </main>
      </ScrollArea>
    </div>
  )
}
