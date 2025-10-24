/**
 * App Mode Configuration
 *
 * This application can run in three distinct modes:
 * 1. registration - For new user registration and onboarding
 * 2. dashboard - Main application dashboard for authenticated users
 * 3. website - Public marketing/documentation website
 */

export type AppMode = 'registration' | 'dashboard' | 'website'

export const APP_MODE = (process.env.APP_MODE || 'dashboard') as AppMode

// Route definitions for each mode
export const MODE_ROUTES = {
  registration: ['/register'],
  dashboard: [
    '/',
    '/dashboard',
    '/workspaces',
    '/environments',
    '/connection-tokens',
    '/admin',
    '/super-admin',
    '/auth',
  ],
  website: ['/', '/docs', '/pricing', '/about', '/contact', '/blog'],
} as const

/**
 * Check if a pathname belongs to the current app mode
 */
export function isRouteAllowedInMode(pathname: string, mode: AppMode): boolean {
  const allowedRoutes = MODE_ROUTES[mode]

  // Check if pathname starts with any of the allowed routes
  return allowedRoutes.some((route) => pathname.startsWith(route))
}

/**
 * Get the appropriate redirect URL for a given mode
 */
export function getRedirectForMode(mode: AppMode): string {
  switch (mode) {
    case 'registration':
      return '/register'
    case 'dashboard':
      return '/dashboard'
    case 'website':
      return '/'
    default:
      return '/'
  }
}

/**
 * Check if current mode allows a specific route
 */
export function validateRouteAccess(pathname: string): {
  allowed: boolean
  redirectTo?: string
} {
  // Allow API routes and static files in all modes
  if (
    pathname.startsWith('/api/') ||
    pathname.startsWith('/_next/') ||
    pathname.startsWith('/static/') ||
    pathname.match(/\.(ico|png|jpg|jpeg|svg|css|js)$/)
  ) {
    return { allowed: true }
  }

  const allowed = isRouteAllowedInMode(pathname, APP_MODE)

  if (!allowed) {
    return {
      allowed: false,
      redirectTo: getRedirectForMode(APP_MODE),
    }
  }

  return { allowed: true }
}
