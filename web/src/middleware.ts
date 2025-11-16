import { auth } from '@/lib/auth'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { validateRouteAccess, APP_MODE } from '@/lib/app-mode'
import { jwtVerify } from 'jose'
import type { Session } from 'next-auth'

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
  // Skip auth checks for auth pages, installation scripts, superadmin login, kltun install, and public assets
  if (
    pathname.startsWith('/auth') ||
    pathname.startsWith('/api') ||
    pathname.startsWith('/install') ||
    pathname.startsWith('/uninstall') ||
    pathname.startsWith('/kltun') ||
    pathname.startsWith('/superadmin-login') ||
    pathname.startsWith('/_next') ||
    pathname.includes('.')
  ) {
    return NextResponse.next()
  }

  // Try to get session from NextAuth
  let session = await auth()
  let userRoles: string[] = []

  // If no NextAuth session, check for superadmin JWT token
  if (!session) {
    const cookieName = process.env.NODE_ENV === 'production'
      ? '__Secure-next-auth.session-token'
      : 'next-auth.session-token'

    const token = req.cookies.get(cookieName)?.value

    if (token) {
      try {
        const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
        const { payload } = await jwtVerify(token, secret)

        // Check if this is a superadmin token
        if (payload.provider === 'superadmin-login' && payload.roles) {
          // Create a mock session for superadmin
          userRoles = payload.roles as string[]
          session = {
            user: {
              email: payload.email as string,
              name: payload.name as string,
              roles: userRoles,
            },
            expires: new Date(payload.exp! * 1000).toISOString(),
          } as Session
        }
      } catch (error) {
        console.error('Failed to verify superadmin token:', error)
      }
    }
  }

  // Redirect to login if not authenticated
  if (!session) {
    return NextResponse.redirect(new URL('/auth/signin', req.url))
  }

  // Get user roles
  userRoles = session?.user?.roles || []
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
