"use client"

import { User, Settings, LogOut, Shield } from "lucide-react"
import { useRouter } from "next/navigation"
import { signOut } from "next-auth/react"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"


interface UserMenuProps {
  user: {
    name?: string | null
    email?: string | null
    image?: string | null
  }
  canManagePlatform?: boolean
}

export function UserMenu({ user, canManagePlatform = false }: UserMenuProps) {
  const router = useRouter()
  const initials = user.name
    ?.split(" ")
    .map(n => n[0])
    .join("")
    .toUpperCase() || user.email?.[0].toUpperCase() || "U"

  return (
    <div className="flex items-center gap-2">
      {/* User Menu */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="relative h-9 w-9 rounded-full">
            <Avatar className="h-9 w-9">
              <AvatarImage src={user.image || undefined} alt={user.name || "User"} />
              <AvatarFallback>{initials}</AvatarFallback>
            </Avatar>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56" align="end" forceMount>
          <DropdownMenuLabel className="font-normal">
            <div className="flex flex-col space-y-1">
              <p className="text-sm font-medium leading-none">{user.name || "User"}</p>
              <p className="text-xs leading-none text-muted-foreground">
                {user.email}
              </p>
            </div>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem>
              <User className="mr-2 h-4 w-4" />
              <span>Profile</span>
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </DropdownMenuItem>
            {canManagePlatform && (
              <DropdownMenuItem onClick={() => router.push('/platform')}>
                <Shield className="mr-2 h-4 w-4" />
                <span>Platform Management</span>
              </DropdownMenuItem>
            )}
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="text-red-600"
            onClick={() => signOut({ callbackUrl: "/" })}
          >
            <LogOut className="mr-2 h-4 w-4" />
            <span>Sign out</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}