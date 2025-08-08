import { approveTeamRequest, rejectTeamRequest } from "@/app/actions/team-approvals"

export type ActionHandler = (data: Record<string, string>) => Promise<{ success: boolean; error?: string }>

export const notificationActionHandlers: Record<string, ActionHandler> = {
  "/api/teams/approve": async (data) => {
    if (!data.requestId) {
      return { success: false, error: "Request ID is required" }
    }
    return approveTeamRequest(data.requestId)
  },
  
  "/api/teams/reject": async (data) => {
    if (!data.requestId) {
      return { success: false, error: "Request ID is required" }
    }
    return rejectTeamRequest(data.requestId, data.reason || "No reason provided")
  },
  
  // Add more action handlers here as needed
  // "/api/invitations/accept": async (data) => { ... },
  // "/api/invitations/decline": async (data) => { ... },
}

export async function executeNotificationAction(endpoint: string, data?: Record<string, string>): Promise<{ success: boolean; error?: string }> {
  const handler = notificationActionHandlers[endpoint]
  
  if (!handler) {
    return { success: false, error: `No handler found for endpoint: ${endpoint}` }
  }
  
  try {
    return await handler(data || {})
  } catch (error) {
    // Error executing action
    return { success: false, error: error instanceof Error ? error.message : "Unknown error" }
  }
}