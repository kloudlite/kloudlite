'use client'

import { LogOut, Shield } from 'lucide-react'
import Link from 'next/link'
import {
  Button,
  Avatar,
  AvatarFallback,
  AvatarImage,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { signOutAction } from '@/app/actions/auth'

interface UserProfileDropdownProps {
  email?: string
  displayName?: string
  isAdmin?: boolean
  isSuperAdmin?: boolean
}

export function UserProfileDropdown({
  email,
  displayName,
  isAdmin = false,
  isSuperAdmin = false,
}: UserProfileDropdownProps) {
  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="relative h-10 gap-2 px-2 hover:bg-muted/50 transition-colors">
          <Avatar className="h-8 w-8 ring-2 ring-foreground/10">
            <AvatarImage src={undefined} alt={displayName || 'User'} />
            <AvatarFallback className="text-xs font-semibold bg-primary/10 text-primary">
              {getInitials(displayName || email || 'User')}
            </AvatarFallback>
          </Avatar>
          <div className="hidden sm:flex flex-col items-start text-left">
            <span className="text-sm font-medium">{displayName || 'User'}</span>
            <span className="text-muted-foreground text-xs">{email}</span>
          </div>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>
          <div className="flex flex-col space-y-1">
            <p className="text-sm leading-none font-semibold">{displayName || 'User'}</p>
            <p className="text-muted-foreground text-xs leading-none">{email}</p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {(isAdmin || isSuperAdmin) && (
          <>
            <DropdownMenuItem asChild className="cursor-pointer">
              <Link href="/admin">
                <Shield className="mr-2 h-4 w-4" />
                Administration
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}
        <DropdownMenuItem
          onClick={() => signOutAction()}
          className="text-red-600 dark:text-red-400 cursor-pointer focus:bg-red-500/10 focus:text-red-600 dark:focus:text-red-400 hover:bg-red-500/10"
        >
          <LogOut className="mr-2 h-4 w-4" />
          Sign Out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
