import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"

import { getAuthOptions } from "@/lib/auth/get-auth-options"
import { getAccountsClient, getAuthMetadata } from "@/lib/grpc/accounts-client"
import { getAuthClient } from "@/lib/grpc/auth-client"

import { TeamLayout } from "./team-layout"

interface LayoutProps {
  children: React.ReactNode
  params: Promise<{ teamSlug: string }>
}

export default async function Layout({ children, params }: LayoutProps) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  const { teamSlug } = await params

  if (!session) {
    redirect(`/auth/login?callbackUrl=/${teamSlug}`)
  }

  if (!session.user.emailVerified) {
    redirect("/auth/email-verification-required")
  }

  const accountsClient = getAccountsClient()
  const authClient = getAuthClient()
  const metadata = await getAuthMetadata()

  // Fetch team details and user's role
  let team = null
  let userRole = "member" // default role
  let canManagePlatform = false
  
  try {
    // Parallel fetch team details and platform role
    const [teamsResponse, platformRole] = await Promise.all([
      // Get teams
      new Promise<any>((resolve, reject) => {
        accountsClient.listTeams({}, metadata, (error, response) => {
          if (error) reject(error)
          else resolve(response)
        })
      }),
      // Get platform role
      new Promise<any>((resolve) => {
        authClient.getPlatformRole({}, metadata, (error, response) => {
          if (error) {
            console.error('Failed to get platform role:', error)
            resolve({ canManagePlatform: false })
          } else {
            resolve(response)
          }
        })
      })
    ])
    
    // Find team by slug
    team = teamsResponse?.teams?.find((t: any) => t.slug === teamSlug)
    
    if (!team) {
      redirect("/overview")
    }

    // Determine user's role in the team
    // For now, we'll assume owner if userId matches team.ownerId
    if (team.ownerId === session.user.id) {
      userRole = "owner"
    }
    // TODO: Check admin list once it's available from backend
    
    // Set platform management permission
    canManagePlatform = platformRole?.canManagePlatform || false
    
  } catch (error) {
    console.error("Error fetching team:", error)
    redirect("/overview")
  }

  return (
    <TeamLayout 
      team={team}
      userRole={userRole}
      session={session}
      teamSlug={teamSlug}
      canManagePlatform={canManagePlatform}
    >
      {children}
    </TeamLayout>
  )
}