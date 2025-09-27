import Link from "next/link"
import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"

import { listUserTeams, listUserPendingTeamRequests } from "@/app/actions/teams"
import { OverviewHeader } from "@/components/overview/overview-header"
import { TeamsList } from "@/components/overview/teams-list"
import { ThemeToggle } from "@/components/theme-toggle"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { getAuthOptions } from "@/lib/auth/get-auth-options"
import { getAuthClient, getAuthMetadata } from "@/lib/grpc/auth-client"

export default async function OverviewPage({
  searchParams,
}: {
  searchParams: Promise<{ teamCreated?: string; teamPending?: string }>
}) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  const { teamCreated, teamPending } = await searchParams

  if (!session) {
    redirect("/auth/login")
  }

  // Parallel fetch teams, pending requests, and platform role
  const [teams, pendingRequests, platformInfo] = await Promise.all([
    // Fetch teams
    listUserTeams().catch(error => {
      console.error('Failed to fetch teams:', error)
      return []
    }),
    
    // Fetch pending team requests
    listUserPendingTeamRequests().catch(error => {
      console.error('Failed to fetch pending team requests:', error)
      return []
    }),
    
    // Check platform admin access and if user can create teams
    (async () => {
      try {
        const client = getAuthClient()
        const metadata = await getAuthMetadata()
        interface PlatformRoleResponse {
          canManagePlatform?: boolean
          canCreateTeams?: boolean
        }
        const platformRole = await new Promise<PlatformRoleResponse>((resolve, reject) => {
          client.getPlatformRole({}, metadata, (error, response) => {
            if (error) {reject(error)}
            else {resolve(response)}
          })
        })
        return {
          canManagePlatform: platformRole?.canManagePlatform || false,
          canCreateTeams: platformRole?.canCreateTeams !== false // Default to true if not specified
        }
      } catch (error) {
        // Silently fail - platform role check is not critical
        return { canManagePlatform: false, canCreateTeams: true }
      }
    })()
  ])

  // Transform teams data to include mock counts for now
  const enrichedTeams = (teams as any[]).map((team) => ({
    accountid: team.teamId,
    name: team.displayName,
    slug: team.slug,
    status: team.status || 'active',
    role: team.role,
    memberCount: 1, // TODO: Fetch actual member count
    resourceCount: Math.floor(Math.random() * 10), // TODO: Fetch actual resource count
    description: team.description || `Resources and infrastructure for ${team.displayName}`,
  }))
  
  // Add pending team requests to the list
  const pendingTeams = (pendingRequests as any[]).map((request) => ({
    accountid: request.requestId,
    name: request.displayName || request.slug,
    slug: request.slug,
    status: 'pending',
    role: 'owner', // User will be owner once approved
    memberCount: 0,
    resourceCount: 0,
    description: request.description || `Pending approval for ${request.slug}`,
    createdAt: request.requestedAt,
  }))
  
  const allTeams = [...enrichedTeams, ...pendingTeams]

  return (
    <div className="relative flex min-h-screen flex-col bg-background">
      {/* Optimized background gradient */}
      <div className="pointer-events-none">
        <div className="absolute inset-0 gradient-primary -z-10" />
      </div>
      
      {/* Header */}
      <OverviewHeader 
        user={session.user} 
        canManagePlatform={platformInfo.canManagePlatform} 
      />
      
      {/* Main Content */}
      <main className="flex-1 container mx-auto px-4 py-6 md:px-6 md:py-8 lg:px-8">
        {/* Page Header */}
        <div className="mb-6 space-y-1 md:mb-8">
          <h1 className="text-2xl font-extralight tracking-tight text-foreground/90 md:text-3xl">Your Teams</h1>
          <p className="text-sm text-muted-foreground md:text-base">
            Manage your teams and access their resources
          </p>
        </div>

        {/* Success/Pending Alerts */}
        {teamCreated && (
          <Alert variant="success" className="mb-6">
            <AlertDescription>
              Your team has been created successfully! You can now access it from the list below.
            </AlertDescription>
          </Alert>
        )}
        
        {teamPending && pendingTeams.length > 0 && (
          <Alert variant="warning" className="mb-6">
            <AlertDescription>
              Your team creation request has been submitted and is pending approval. You'll be notified once it's approved.
            </AlertDescription>
          </Alert>
        )}

        {/* Teams List */}
        <TeamsList 
          teams={allTeams} 
          canCreateTeam={platformInfo.canCreateTeams}
        />
      </main>
      
      {/* Footer */}
      <footer className="relative mt-auto">
        <div className="absolute inset-x-0 top-0 h-px bg-border/50" />
        <div className="container mx-auto px-4 py-6 md:px-6 md:py-8">
          <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
            <nav className="flex flex-wrap items-center justify-center gap-4 text-xs md:gap-6 md:text-sm">
              <Link href="#" className="text-muted-foreground transition-colors hover:text-foreground">
                Documentation
              </Link>
              <Link href="#" className="text-muted-foreground transition-colors hover:text-foreground">
                API Reference
              </Link>
              <Link href="#" className="text-muted-foreground transition-colors hover:text-foreground">
                Support
              </Link>
              <Link href="/status" className="text-muted-foreground transition-colors hover:text-foreground">
                Status
              </Link>
              <Link href="/legal/privacy" className="text-muted-foreground transition-colors hover:text-foreground">
                Privacy
              </Link>
              <Link href="/legal/terms" className="text-muted-foreground transition-colors hover:text-foreground">
                Terms
              </Link>
            </nav>
            
            <div className="flex items-center gap-4 md:gap-6">
              <span className="text-xs text-muted-foreground md:text-sm">© 2024 Kloudlite</span>
              <ThemeToggle />
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}