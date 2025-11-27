import Link from 'next/link'
import { getSession } from '@/lib/get-session'
import { signOutAction } from '@/app/actions/auth'
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { ChevronDown, User, LogOut, Home } from 'lucide-react'
import { AdminNavigation } from './_components/admin-navigation'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { isSystemReady, SystemSetupPage } from '@/lib/system-check'

// Admin layout - middleware ensures only users with admin/super-admin roles (and no 'user' role) can access this
export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession()

  // Session and role access is guaranteed by middleware
  const userRoles = session!.user?.roles || []
  const hasUserRole = userRoles.includes('user')
  const isSuperAdmin = userRoles.includes('super-admin')

  // Check if system is configured
  const systemReady = await isSystemReady()

  // If system not ready and not super-admin, show under construction
  if (!systemReady && !isSuperAdmin) {
    return <SystemSetupPage />
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Admin Header */}
      <header className="border-b bg-white">
        <div className="mx-auto max-w-7xl px-6">
          <div className="flex h-16 items-center justify-between">
            {/* Logo / Brand and Navigation */}
            <div className="flex items-center gap-8">
              <div className="flex items-center gap-3">
                <KloudliteLogo className="text-lg font-medium" />
                <span className="text-lg font-medium text-gray-600">Admin</span>
              </div>

              {/* Admin Navigation */}
              <AdminNavigation />
            </div>

            {/* User Dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="gap-1">
                  <User className="h-4 w-4" />
                  <span className="hidden sm:inline">{session!.user?.name || 'User'}</span>
                  <ChevronDown className="h-3 w-3" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium">{session!.user?.name || 'User'}</p>
                    <p className="text-xs text-gray-500">{session!.user?.email}</p>
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
                <form action={signOutAction}>
                  <DropdownMenuItem variant="destructive" asChild>
                    <button type="submit" className="w-full">
                      <LogOut className="mr-2 h-4 w-4" />
                      Sign out
                    </button>
                  </DropdownMenuItem>
                </form>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main>{children}</main>
    </div>
  )
}
