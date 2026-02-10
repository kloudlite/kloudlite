'use client'

import Link from 'next/link'
import { ChevronDown, User, LogOut, Home } from 'lucide-react'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { signOutAction } from '@/app/actions/auth'

interface AdminProfileDropdownProps {
  name?: string | null
  email?: string | null
  hasUserRole: boolean
}

export function AdminProfileDropdown({ name, email, hasUserRole }: AdminProfileDropdownProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-1">
          <User className="h-4 w-4" />
          <span className="hidden sm:inline">{name || 'User'}</span>
          <ChevronDown className="h-3 w-3" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel className="font-normal">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium">{name || 'User'}</p>
            <p className="text-muted-foreground text-xs">{email}</p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {hasUserRole && (
          <>
            <DropdownMenuItem asChild>
              <Link href="/" className="cursor-pointer">
                <Home className="mr-2 h-4 w-4" />
                Dashboard
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}
        <DropdownMenuItem
          onClick={() => signOutAction()}
          className="text-destructive focus:text-destructive cursor-pointer"
        >
          <LogOut className="mr-2 h-4 w-4" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
