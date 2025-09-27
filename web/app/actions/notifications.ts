"use server"

import * as grpc from "@grpc/grpc-js"
import { getServerSession } from "next-auth"

import { getAuthOptions } from "@/lib/auth/get-auth-options"
import { getAuthClient } from "@/lib/auth/grpc-client"

interface NotificationListParams {
  limit?: number
  offset?: number
  unreadOnly?: boolean
  actionRequiredOnly?: boolean
}

export async function listNotifications({
  limit = 50,
  offset = 0,
  unreadOnly = false,
  actionRequiredOnly = false,
}: NotificationListParams) {
  const session = await getServerSession(await getAuthOptions())
  if (!session?.user?.id || !session?.accessToken) {
    throw new Error("Unauthorized")
  }

  const client = getAuthClient()
  const metadata = new grpc.Metadata()
  metadata.add("authorization", `Bearer ${session.accessToken}`)

  return new Promise((resolve, reject) => {
    client.listNotifications(
      {
        limit,
        offset,
        unreadOnly,
        actionRequiredOnly,
      },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          resolve({
            notifications: response?.notifications || [],
            totalCount: response?.totalCount || 0,
          })
        }
      }
    )
  })
}

export async function getUnreadNotificationCount(): Promise<number> {
  const session = await getServerSession(await getAuthOptions())
  if (!session?.user?.id || !session?.accessToken) {
    return 0
  }

  try {
    const client = getAuthClient()
    const metadata = new grpc.Metadata()
    metadata.add("authorization", `Bearer ${session.accessToken}`)

    return new Promise((resolve) => {
      client.getUnreadNotificationCount(
        {},
        metadata,
        (error, response) => {
          if (error) {
            console.error("Failed to get unread notification count:", error)
            resolve(0)
          } else {
            resolve(response?.count || 0)
          }
        }
      )
    })
  } catch (error) {
    console.error("Failed to get unread notification count:", error)
    return 0
  }
}

export async function markNotificationAsRead(notificationId: string) {
  const session = await getServerSession(await getAuthOptions())
  if (!session?.user?.id || !session?.accessToken) {
    throw new Error("Unauthorized")
  }

  const client = getAuthClient()
  const metadata = new grpc.Metadata()
  metadata.add("authorization", `Bearer ${session.accessToken}`)

  return new Promise<void>((resolve, reject) => {
    client.markNotificationAsRead(
      { notificationId },
      metadata,
      (error) => {
        if (error) {
          reject(error)
        } else {
          resolve()
        }
      }
    )
  })
}


export async function markNotificationActionTaken(notificationId: string, actionId: string) {
  const session = await getServerSession(await getAuthOptions())
  if (!session?.user?.id || !session?.accessToken) {
    throw new Error("Unauthorized")
  }

  const client = getAuthClient()
  const metadata = new grpc.Metadata()
  metadata.add("authorization", `Bearer ${session.accessToken}`)

  return new Promise<void>((resolve, reject) => {
    client.markNotificationActionTaken(
      { notificationId, actionId },
      metadata,
      (error) => {
        if (error) {
          reject(error)
        } else {
          resolve()
        }
      }
    )
  })
}