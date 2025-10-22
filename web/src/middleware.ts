import { auth } from '@/lib/auth'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { validateRouteAccess, APP_MODE } from '@/lib/app-mode'

export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl

  // Enforce API route mode restrictions
  if (pathname.startsWith('/api/register/') && APP_MODE !== 'registration') {
    return NextResponse.json(
      { error: 'Registration API routes are only available in registration mode' },
      { status: 403 }
    )
  }

  // First, enforce app mode routing at the top level
  const routeAccess = validateRouteAccess(pathname)
  if (!routeAccess.allowed && routeAccess.redirectTo) {
    return NextResponse.redirect(new URL(routeAccess.redirectTo, req.url))
  }

  // Mode-specific middleware logic
  switch (APP_MODE) {
    case 'registration':
      return handleRegistrationMode(req, pathname)
    case 'dashboard':
      return handleDashboardMode(req, pathname)
    case 'website':
      return handleWebsiteMode(req, pathname)
    default:
      return NextResponse.next()
  }
}

/**
 * Registration mode middleware
 * Handles registration flow authentication
 */
async function handleRegistrationMode(
  req: NextRequest,
  pathname: string
): Promise<NextResponse> {
  // Skip middleware for API routes, OAuth callbacks, signout, and public assets
  if (
    pathname.startsWith('/api/') ||
    pathname.startsWith('/_next') ||
    pathname.includes('.') ||
    pathname === '/register/signout' ||
    pathname.startsWith('/register/oauth/')
  ) {
    return NextResponse.next()
  }

  // Allow /register page without session
  if (pathname === '/register') {
    return NextResponse.next()
  }

  // For all other /register/* routes, validate the registration flow
  if (pathname.startsWith('/register/')) {
    try {
      // Get registration session cookie
      const sessionCookie = req.cookies.get('registration_session')

      if (!sessionCookie) {
        // No session, redirect to register page
        return NextResponse.redirect(new URL('/register', req.url))
      }

      // Decode JWT to get installationKey and user email
      const { jwtVerify } = await import('jose')
      const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
      const { payload } = await jwtVerify(sessionCookie.value, secret)
      const installationKey = payload.installationKey as string
      const email = payload.email as string

      if (!installationKey || !email) {
        return NextResponse.redirect(new URL('/register', req.url))
      }

      // Get user status directly from storage (DO NOT call verify-key as it generates secret)
      const { getUserByEmail } = await import('@/lib/registration/storage-service')
      const user = await getUserByEmail(email)

      if (!user) {
        // User not found
        return NextResponse.redirect(new URL('/register', req.url))
      }

      // Determine correct page based on state
      if (!user.hasCompletedInstallation) {
        // Installation not complete - should be on install page
        if (pathname !== '/register/install') {
          return NextResponse.redirect(new URL('/register/install', req.url))
        }
      } else if (!user.subdomain) {
        // Installation complete but no subdomain - should be on domain page
        if (pathname !== '/register/domain') {
          return NextResponse.redirect(new URL('/register/domain', req.url))
        }
      } else {
        // Subdomain assigned - should be on complete page
        if (pathname !== '/register/complete') {
          return NextResponse.redirect(new URL('/register/complete', req.url))
        }
      }

      return NextResponse.next()
    } catch (error) {
      console.error('Registration middleware error:', error)
      return NextResponse.redirect(new URL('/register', req.url))
    }
  }

  return NextResponse.next()
}

/**
 * Dashboard mode middleware
 * Handles main application authentication and role-based access
 */
async function handleDashboardMode(
  req: NextRequest,
  pathname: string
): Promise<NextResponse> {
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
  const hasAdminRole =
    userRoles.includes('admin') || userRoles.includes('super-admin')

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
async function handleWebsiteMode(
  req: NextRequest,
  pathname: string
): Promise<NextResponse> {
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
