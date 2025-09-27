"use client"

import { useState, useEffect } from "react"

import { Bell } from "lucide-react"

import { getUnreadNotificationCount } from "@/app/actions/notifications"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetTrigger,
  SheetTitle,
} from "@/components/ui/sheet"

import { NotificationList } from "./notification-list"


export function NotificationBell() {
  const [isOpen, setIsOpen] = useState(false)
  const [unreadCount, setUnreadCount] = useState(0)

  // Fetch unread count
  useEffect(() => {
    const fetchUnreadCount = async () => {
      try {
        const count = await getUnreadNotificationCount()
        setUnreadCount(count)
      } catch (_error) {
        // Silently fail - notification count is not critical
      }
    }

    fetchUnreadCount()
    
    // Refresh count when sheet closes
    if (!isOpen) {
      fetchUnreadCount()
    }
  }, [isOpen])

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <span className="absolute top-0 right-0 h-2 w-2 rounded-full bg-red-500" />
          )}
        </Button>
      </SheetTrigger>
      <SheetContent className="w-full sm:w-[400px] p-0 bg-background">
        <SheetTitle className="sr-only">Notifications</SheetTitle>
        <NotificationList onClose={() => setIsOpen(false)} />
      </SheetContent>
    </Sheet>
  )
}