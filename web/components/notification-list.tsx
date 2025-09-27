"use client"

import { useState, useEffect } from "react"

import { formatDistanceToNow } from "date-fns"
import { UserPlus, Mail, AlertCircle, Bell, Loader2, X } from "lucide-react"

import {
  listNotifications,
  markNotificationAsRead,
  markNotificationActionTaken,
} from "@/app/actions/notifications"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"



interface NotificationAction {
  id: string
  label: string
  style: string
  endpoint: string
  method: string
  data?: Record<string, string>
}

interface Notification {
  id: string
  title: string
  description: string
  type: string
  targetType?: string
  actionRequired: boolean
  actions?: NotificationAction[]
  createdAt: string
  read: boolean
  actionTaken?: boolean
}

interface NotificationListProps {
  onClose?: () => void
}

export function NotificationList({ onClose }: NotificationListProps) {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [loading, setLoading] = useState(true)

  const fetchNotifications = async () => {
    setLoading(true)
    try {
      const data = await listNotifications({
        limit: 50,
        offset: 0,
      })
      // Sort notifications by newest first
      const sortedNotifications = [...data.notifications].sort((a, b) => 
        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      )
      setNotifications(sortedNotifications)
      
      // Mark all unread notifications as read
      const unreadNotifications = sortedNotifications.filter(n => !n.read)
      unreadNotifications.forEach(notification => {
        handleMarkAsRead(notification.id)
      })
    } catch (_error) {
      // Failed to fetch notifications
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchNotifications()
  }, [])

  const handleMarkAsRead = async (notificationId: string) => {
    try {
      await markNotificationAsRead(notificationId)
      setNotifications((prev) =>
        prev.map((n) =>
          n.id === notificationId ? { ...n, read: true } : n
        )
      )
    } catch (_error) {
      // Failed to mark notification as read
    }
  }

  const handleActionTaken = async (notificationId: string, actionId: string) => {
    try {
      await markNotificationActionTaken(notificationId, actionId)
      // Refresh notifications to update UI
      fetchNotifications()
    } catch (_error) {
      // Failed to mark action as taken
    }
  }

  const handleAction = async (notification: Notification, action: NotificationAction) => {
    try {
      if (action.method === "GET") {
        // For GET requests, navigate to the endpoint
        await handleActionTaken(notification.id, action.id)
        window.location.href = action.endpoint
      } else {
        // For server actions
        const { executeNotificationAction } = await import("@/lib/notification-actions")
        const result = await executeNotificationAction(action.endpoint, action.data)
        
        if (result.success) {
          await handleActionTaken(notification.id, action.id)
          fetchNotifications()
          if (onClose) {onClose()}
        } else {
          throw new Error(result.error || "Action failed")
        }
      }
    } catch (_error) {
      // Failed to execute action
      // TODO: Show error toast
    }
  }


  const getNotificationIcon = (type: string) => {
    switch (type) {
      case "team_request":
      case "team_approved":
      case "team_rejected":
        return <UserPlus className="h-4 w-4" />
      case "email_verification":
        return <Mail className="h-4 w-4" />
      case "system":
      case "announcement":
        return <AlertCircle className="h-4 w-4" />
      default:
        return <Bell className="h-4 w-4" />
    }
  }

  const getIconColorClass = (type: string) => {
    switch (type) {
      case "team_request":
      case "team_approved":
      case "team_rejected":
        return "bg-blue-500/10 text-blue-500"
      case "email_verification":
        return "bg-orange-500/10 text-orange-500"
      case "system":
      case "announcement":
        return "bg-muted text-muted-foreground"
      default:
        return "bg-green-500/10 text-green-500"
    }
  }

  const unreadCount = notifications.filter((n) => !n.read).length

  return (
    <div className="flex h-full flex-col p-0">
      <div className="flex h-16 items-center justify-between border-b px-4 pr-12 sm:px-6">
        <div>
          <h2 className="text-lg font-semibold">Notifications</h2>
          <p className="text-xs text-muted-foreground">
            You have {unreadCount} unread notification{unreadCount !== 1 ? 's' : ''}
          </p>
        </div>
      </div>
      
      <ScrollArea className="flex-1">
        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Bell className="h-12 w-12 text-muted-foreground/20 mb-4" />
            <p className="text-sm text-muted-foreground">No notifications yet</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {notifications.map((notification) => (
              <div
                key={notification.id}
                className={cn(
                  "px-4 sm:px-6 py-4 transition-colors hover:bg-muted/30",
                  !notification.read && "bg-muted/10"
                )}
              >
                <div className="flex gap-3">
                  <div className={cn(
                    "mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
                    getIconColorClass(notification.type)
                  )}>
                    {getNotificationIcon(notification.type)}
                  </div>
                  
                  <div className="flex-1 space-y-1 min-w-0">
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0 flex-1">
                        <p className={cn(
                          "text-sm leading-none",
                          !notification.read ? "font-semibold" : "font-medium"
                        )}>
                          {notification.title}
                        </p>
                      </div>
                      <p className="text-xs text-muted-foreground whitespace-nowrap">
                        {formatDistanceToNow(new Date(notification.createdAt), { addSuffix: true })}
                      </p>
                    </div>
                    
                    <p className={cn(
                      "text-xs text-muted-foreground leading-relaxed",
                      !notification.read && "text-foreground/80"
                    )}>
                      {notification.description}
                    </p>
                    
                    {notification.actionRequired && notification.actions && notification.actions.length > 0 && !notification.actionTaken && (
                      <div className="flex flex-wrap gap-2 pt-2">
                        {notification.actions.map((action) => (
                          <Button
                            key={action.id}
                            size="sm"
                            variant={action.style === "danger" ? "destructive" : action.style === "primary" ? "default" : "outline"}
                            onClick={(e) => {
                              e.stopPropagation()
                              handleAction(notification, action)
                            }}
                            className="h-7 px-3 text-xs"
                          >
                            {action.label}
                          </Button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}