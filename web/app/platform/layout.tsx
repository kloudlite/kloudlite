import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"

import { getAuthOptions } from "@/lib/auth/get-auth-options"
import { getAccountsClient, getAuthMetadata } from "@/lib/grpc/accounts-client"
import { getAuthClient } from "@/lib/grpc/auth-client"

import { PlatformLayout } from "./platform-layout"

export default async function Layout({
  children,
}: {
  children: React.ReactNode
}) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)

  if (!session) {
    redirect("/auth/login")
  }

  const accountsClient = getAccountsClient()
  const accountsMetadata = await getAuthMetadata()
  const authClient = getAuthClient()
  const authMetadata = await getAuthMetadata()

  // Parallel fetch platform role and team requests count
  const [platformRole, teamRequests] = await Promise.all([
    // Get platform role
    new Promise<any>((resolve, reject) => {
      authClient.getPlatformRole({}, authMetadata, (error, response) => {
        if (error) {reject(error)}
        else {resolve(response)}
      })
    }),
    
    // Get team requests count
    new Promise<any>((resolve, reject) => {
      accountsClient.listTeamRequests({ status: 'pending' }, accountsMetadata, (error, response) => {
        if (error) {reject(error)}
        else {resolve(response?.requests || [])}
      })
    })
  ])

  if (!platformRole?.canManagePlatform) {
    redirect("/overview")
  }

  return (
    <PlatformLayout 
      platformRole={platformRole}
      session={session}
      teamRequestsCount={teamRequests.length}
    >
      {children}
    </PlatformLayout>
  )
}