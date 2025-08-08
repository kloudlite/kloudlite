"use server"

import * as grpc from "@grpc/grpc-js"
import { revalidatePath } from "next/cache"
import { getServerSession } from "next-auth"

import { getAuthOptions } from "@/lib/auth/get-auth-options"
import { getAccountsClient } from "@/lib/grpc/accounts-client"

export async function approveTeamRequest(requestId: string) {
  const session = await getServerSession(await getAuthOptions())
  if (!session?.user?.id || !session?.accessToken) {
    throw new Error("Unauthorized")
  }

  const client = getAccountsClient()
  const metadata = new grpc.Metadata()
  metadata.add("authorization", `Bearer ${session.accessToken}`)

  return new Promise<{ success: boolean; teamId?: string; error?: string }>((resolve) => {
    client.approveTeamRequest(
      { requestId },
      metadata,
      (error, response) => {
        if (error) {
          console.error("Failed to approve team request:", error)
          resolve({ success: false, error: error.message })
        } else {
          // Revalidate relevant pages
          revalidatePath("/platform/teams/requests")
          revalidatePath("/overview")
          
          resolve({ 
            success: true, 
            teamId: response?.teamId 
          })
        }
      }
    )
  })
}

export async function rejectTeamRequest(requestId: string, reason: string) {
  const session = await getServerSession(await getAuthOptions())
  if (!session?.user?.id || !session?.accessToken) {
    throw new Error("Unauthorized")
  }

  const client = getAccountsClient()
  const metadata = new grpc.Metadata()
  metadata.add("authorization", `Bearer ${session.accessToken}`)

  return new Promise<{ success: boolean; error?: string }>((resolve) => {
    client.rejectTeamRequest(
      { requestId, reason },
      metadata,
      (error, response) => {
        if (error) {
          console.error("Failed to reject team request:", error)
          resolve({ success: false, error: error.message })
        } else {
          // Revalidate relevant pages
          revalidatePath("/platform/teams/requests")
          
          resolve({ success: response?.success || false })
        }
      }
    )
  })
}