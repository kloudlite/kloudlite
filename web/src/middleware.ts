import { NextRequest, NextResponse } from 'next/server'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Public routes that don't require authentication
  const publicRoutes = [
    '/',
    '/auth/login',
    '/auth/signup',
    '/auth/forgot-password',
    '/auth/reset-password',
    '/auth/verify-email',
    '/auth/error',
    '/auth/callback',
    '/privacy',
    '/terms',
    '/sso-demo',
  ]

  // Auth routes that logged-in users shouldn't access
  const authRoutes = [
    '/auth/login',
    '/auth/signup',
    '/auth/forgot-password',
    '/auth/reset-password',
    '/auth/verify-email',
  ]

  // Skip middleware for API routes, static files, and NextAuth
  if (
    pathname.startsWith('/api/') ||
    pathname.startsWith('/_next/') ||
    pathname.startsWith('/favicon.ico')
  ) {
    return NextResponse.next()
  }
  
  // IMPORTANT: Skip middleware for NextAuth callback URLs
  if (pathname.startsWith('/api/auth/')) {
    return NextResponse.next()
  }

  // Check if user has auth token (basic check)
  const authToken = request.cookies.get('auth-token')
  // NextAuth might chunk the session cookie, so check for the base cookie or chunked versions
  const nextAuthToken = request.cookies.get('next-auth.session-token') || 
                       request.cookies.get('__Secure-next-auth.session-token') ||
                       request.cookies.get('next-auth.session-token.0') || // Chunked cookie
                       request.cookies.get('__Secure-next-auth.session-token.0') // Secure chunked cookie
  const isLoggedIn = !!(authToken || nextAuthToken)

  const isPublicRoute = publicRoutes.includes(pathname)
  const isAuthRoute = authRoutes.includes(pathname)

  // If user is logged in and trying to access auth routes, redirect to teams
  if (isLoggedIn && isAuthRoute) {
    return NextResponse.redirect(new URL('/teams', request.url))
  }

  // If user is not logged in and trying to access protected routes, redirect to login
  if (!isLoggedIn && !isPublicRoute) {
    return NextResponse.redirect(new URL('/auth/login', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}