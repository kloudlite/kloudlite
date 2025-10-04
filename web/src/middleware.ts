import { auth } from '@/lib/auth'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function middleware(req: NextRequest) {
  // Get session from auth
  const session = await auth()
  const { pathname } = req.nextUrl

  // Skip auth checks for auth pages and public assets
  if (pathname.startsWith('/auth') ||
      pathname.startsWith('/api') ||
      pathname.startsWith('/_next') ||
      pathname.includes('.')) {
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
    // Super-admins and admins can access admin section regardless of user role
  } else if (pathname === '/' || pathname.startsWith('/(main)')) {
    // Main dashboard section - redirect admin-only users to admin
    if (!hasUserRole && hasAdminRole) {
      return NextResponse.redirect(new URL('/admin', req.url))
    }
    // Require user role for main section (unless they're admin accessing via direct navigation)
    if (!hasUserRole && !hasAdminRole) {
      return NextResponse.redirect(new URL('/auth/signin', req.url))
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}