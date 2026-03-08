'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Button, Avatar, AvatarFallback, AvatarImage, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger, KloudliteLogo, ThemeSwitcher } from '@kloudlite/ui'
import { Settings, LogOut, UserCircle } from 'lucide-react'
import { OrgSwitcher } from '@/components/org-switcher'

interface Org {
  id: string
  name: string
  slug: string
}

interface InstallationsHeaderProps {
  user: {
    name: string
    email: string
    image?: string
  }
  orgs?: Org[]
  currentOrgId?: string
  installationName?: string
  installationDomain?: string
}

export function InstallationsHeader({ user, orgs, currentOrgId }: InstallationsHeaderProps) {
  const router = useRouter()

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  const handleLogout = async () => {
    // Clear the registration session cookie
    await fetch('/api/register/logout', { method: 'POST' })
    router.push('/login')
  }

  return (
    <header className="bg-background border-b border-foreground/10 sticky top-0 z-50 backdrop-blur-sm bg-background/95">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
        {/* Logo */}
        <div className="flex items-center gap-3">
          <Link href="/installations" className="flex items-center gap-3 transition-opacity hover:opacity-80">
            <KloudliteLogo className="h-7 sm:h-8" linkToHome={false} />
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground text-sm font-light">/</span>
              <span className="text-foreground text-sm font-bold tracking-wide">console</span>
            </div>
          </Link>
          {orgs && currentOrgId && (
            <>
              <span className="text-muted-foreground text-sm font-light">/</span>
              <OrgSwitcher orgs={orgs} currentOrgId={currentOrgId} />
            </>
          )}
        </div>

        {/* Theme & User Menu */}
        <div className="flex items-center gap-1">
          <ThemeSwitcher />
        <DropdownMenu modal={false}>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-10 gap-2 px-2 hover:bg-muted/50 transition-colors">
              <Avatar className="h-8 w-8 ring-2 ring-foreground/10">
                <AvatarImage src={user.image} alt={user.name} />
                <AvatarFallback className="text-xs font-semibold bg-primary/10 text-primary">{getInitials(user.name)}</AvatarFallback>
              </Avatar>
              <div className="hidden sm:flex flex-col items-start text-left">
                <span className="text-sm font-medium">{user.name}</span>
                <span className="text-muted-foreground text-xs">{user.email}</span>
              </div>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel>
              <div className="flex flex-col space-y-1">
                <p className="text-sm leading-none font-semibold">{user.name}</p>
                <p className="text-muted-foreground text-xs leading-none">{user.email}</p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild className="cursor-pointer">
              <Link href="/profile">
                <UserCircle className="mr-2 h-4 w-4" />
                Profile
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild className="cursor-pointer">
              <Link href="/installations/settings">
                <Settings className="mr-2 h-4 w-4" />
                Settings
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleLogout}
              className="text-red-600 dark:text-red-400 cursor-pointer focus:bg-red-500/10 focus:text-red-600 dark:focus:text-red-400 hover:bg-red-500/10"
            >
              <LogOut className="mr-2 h-4 w-4" />
              Sign Out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
