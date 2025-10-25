import { auth } from '@/lib/auth'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { validateRouteAccess, APP_MODE } from '@/lib/app-mode'

export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl

  // First, enforce app mode routing at the top level
  const routeAccess = validateRouteAccess(pathname)
  if (!routeAccess.allowed && routeAccess.redirectTo) {
    return NextResponse.redirect(new URL(routeAccess.redirectTo, req.url))
  }

  // Mode-specific middleware logic
  switch (APP_MODE) {
    case 'dashboard':
      return handleDashboardMode(req, pathname)
    case 'website':
      return handleWebsiteMode(req, pathname)
    default:
      return NextResponse.next()
  }
}

/**
 * Dashboard mode middleware
 * Handles tenant workspace management authentication and role-based access
 * This is for users inside their Kloudlite installation managing workspaces/workmachines
 */
async function handleDashboardMode(req: NextRequest, pathname: string): Promise<NextResponse> {
  const session = await auth()

  // Skip auth checks for auth pages and public assets
  if (
    pathname.startsWith('/auth') ||
    pathname.startsWith('/api') ||
    pathname.startsWith('/_next') ||
    pathname.includes('.')
  ) {
    return NextResponse.next()
  }

  // Redirect to login if not authenticated
  if (!session) {
    return NextResponse.redirect(new URL('/auth/signin', req.url))
  }

  // Get user roles
  const userRoles = session?.user?.roles || []
  const hasUserRole = userRoles.includes('user')
  const hasAdminRole = userRoles.includes('admin') || userRoles.includes('super-admin')

  // Role-based routing logic
  if (pathname.startsWith('/admin')) {
    // Admin section - only allow admin/super-admin access
    if (!hasAdminRole) {
      return NextResponse.redirect(new URL('/', req.url))
    }
  } else if (pathname === '/' || pathname.startsWith('/(main)')) {
    // Main dashboard section - redirect admin-only users to admin
    if (!hasUserRole && hasAdminRole) {
      return NextResponse.redirect(new URL('/admin', req.url))
    }
    // Require user role for main section
    if (!hasUserRole && !hasAdminRole) {
      return NextResponse.redirect(new URL('/auth/signin', req.url))
    }
  }

  return NextResponse.next()
}

/**
 * Website mode middleware
 * Handles public website routing
 */
async function handleWebsiteMode(_req: NextRequest, _pathname: string): Promise<NextResponse> {
  // Website is public, no authentication required
  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
}
