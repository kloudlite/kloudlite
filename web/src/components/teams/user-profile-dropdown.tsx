'use client'

import { useState } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { signOut } from 'next-auth/react'
import { IconButton } from '@/components/ui/icon-button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { User, Settings, LogOut, Shield, Bell, HelpCircle } from 'lucide-react'

export function UserProfileDropdown() {
  const { user, isLoading } = useAuth()
  const [isOpen, setIsOpen] = useState(false)

  const handleLogout = async () => {
    try {
      await signOut({ callbackUrl: '/auth/login' })
    } catch (error) {
      console.error('Logout error:', error)
    }
  }

  if (isLoading) {
    return (
      <div className="w-8 h-8 bg-muted rounded-full animate-pulse" />
    )
  }

  if (!user) {
    return null
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <IconButton variant="ghost" size="sm" className="p-1 rounded-full">
          <Avatar className="h-8 w-8">
            <AvatarImage 
              src={`https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(user.name)}`} 
              alt={user.name}
              className="object-cover" 
            />
            <AvatarFallback>
              {getInitials(user.name)}
            </AvatarFallback>
          </Avatar>
        </IconButton>
      </DropdownMenuTrigger>
      
      <DropdownMenuContent align="end" className="w-72">
        <div className="px-3 py-4">
          <div className="flex items-center gap-3">
            <Avatar className="h-12 w-12">
              <AvatarImage 
                src={`https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(user.name)}`} 
                alt={user.name} 
              />
              <AvatarFallback className="text-base font-medium">
                {getInitials(user.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col min-w-0">
              <span className="font-semibold text-base truncate">
                {user.name}
              </span>
              <span className="text-sm text-muted-foreground truncate">
                {user.email}
              </span>
              {user.verified && (
                <div className="flex items-center gap-1.5 mt-1.5">
                  <div className="flex items-center gap-1 text-xs text-emerald-600 bg-emerald-50 dark:bg-emerald-900/20 px-2 py-0.5 rounded-full">
                    <Shield className="h-3 w-3" />
                    <span className="font-medium">Verified</span>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
        
        <DropdownMenuSeparator />
        
        <div className="py-1">
          <DropdownMenuItem className="gap-3 px-3 py-2.5 cursor-pointer">
            <User className="h-4 w-4" />
            <span className="text-sm font-medium">Profile Settings</span>
          </DropdownMenuItem>
          
          <DropdownMenuItem className="gap-3 px-3 py-2.5 cursor-pointer">
            <Settings className="h-4 w-4" />
            <span className="text-sm font-medium">Account Settings</span>
          </DropdownMenuItem>
          
          <DropdownMenuItem className="gap-3 px-3 py-2.5 cursor-pointer">
            <Bell className="h-4 w-4" />
            <span className="text-sm font-medium">Notifications</span>
          </DropdownMenuItem>
        </div>
        
        <DropdownMenuSeparator />
        
        <div className="py-1">
          <DropdownMenuItem className="gap-3 px-3 py-2.5 cursor-pointer">
            <HelpCircle className="h-4 w-4" />
            <span className="text-sm font-medium">Help & Support</span>
          </DropdownMenuItem>
        </div>
        
        <DropdownMenuSeparator />
        
        <div className="py-1">
          <DropdownMenuItem onClick={handleLogout} variant="destructive" className="gap-3 px-3 py-2.5 cursor-pointer">
            <LogOut className="h-4 w-4" />
            <span className="text-sm font-medium">Sign Out</span>
          </DropdownMenuItem>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}