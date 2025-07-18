'use client'

import { LogoutButton } from '@/components/auth/logout-button'
import { IconButton } from '@/components/ui/icon-button'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { User, Settings, LogOut, Shield, Bell, HelpCircle, Crown, UserCheck, Users } from 'lucide-react'

interface UserProfileDropdownProps {
  variant?: 'default' | 'sidebar'
  userRole?: string
  user: {
    name: string
    email: string
    image?: string
  }
}

export function UserProfileDropdown({ variant = 'default', userRole = 'member', user }: UserProfileDropdownProps) {
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

  const getRoleIcon = (role: string) => {
    switch (role.toLowerCase()) {
      case 'owner':
        return <Crown className="size-3" />
      case 'admin':
        return <Shield className="size-3" />
      case 'member':
      default:
        return <UserCheck className="size-3" />
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        {variant === 'sidebar' ? (
          <Button 
            variant="ghost" 
            className="flex items-center gap-3 w-full px-4 py-4 rounded-lg border text-sm font-medium transition-colors bg-card hover:bg-dashboard-hover hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 h-auto justify-start shadow-dashboard-card-shadow"
            disableActiveTransition={true}
          >
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
            <div className="flex-1 text-left min-w-0">
              <div className="text-sm font-medium truncate">{user.name}</div>
              <div className="text-xs text-muted-foreground truncate">{user.email}</div>
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground/80 truncate">
                {getRoleIcon(userRole)}
                <span className="capitalize">{userRole}</span>
              </div>
            </div>
          </Button>
        ) : (
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
        )}
      </DropdownMenuTrigger>
      
      <DropdownMenuContent align="end" className="w-72">
        {/* Profile Preview Header */}
        <div className="px-4 py-4 border-b border-border">
          <div className="flex items-center gap-3">
            <Avatar className="h-12 w-12">
              <AvatarImage 
                src={`https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(user.name)}`} 
                alt={user.name}
                className="object-cover" 
              />
              <AvatarFallback>
                {getInitials(user.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="text-sm font-semibold truncate">{user.name}</div>
              <div className="text-xs text-muted-foreground truncate">{user.email}</div>
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground/80 mt-1">
                {getRoleIcon(userRole)}
                <span className="capitalize">{userRole}</span>
              </div>
            </div>
          </div>
        </div>
        
        <div className="py-1">
          <DropdownMenuItem className="gap-3 px-4 py-2.5 cursor-pointer">
            <User className="h-4 w-4" />
            <span className="text-sm font-medium">Profile Settings</span>
          </DropdownMenuItem>
          
          <DropdownMenuItem className="gap-3 px-4 py-2.5 cursor-pointer">
            <Settings className="h-4 w-4" />
            <span className="text-sm font-medium">Account Settings</span>
          </DropdownMenuItem>
          
          <DropdownMenuItem className="gap-3 px-4 py-2.5 cursor-pointer">
            <Bell className="h-4 w-4" />
            <span className="text-sm font-medium">Notifications</span>
          </DropdownMenuItem>
        </div>
        
        <DropdownMenuSeparator />
        
        <div className="py-1">
          <DropdownMenuItem className="gap-3 px-4 py-2.5 cursor-pointer">
            <HelpCircle className="h-4 w-4" />
            <span className="text-sm font-medium">Help & Support</span>
          </DropdownMenuItem>
        </div>
        
        <DropdownMenuSeparator />
        
        <div className="py-1">
          <LogoutButton asChild>
            <DropdownMenuItem variant="destructive" className="gap-3 px-3 py-2.5 cursor-pointer">
              <LogOut className="h-4 w-4" />
              <span className="text-sm font-medium">Sign Out</span>
            </DropdownMenuItem>
          </LogoutButton>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}