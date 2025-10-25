/**
 * App Mode Configuration
 *
 * This application can run in two distinct modes:
 * 1. website - Public marketing/documentation website + installation console
 * 2. dashboard - Tenant's workspace management (inside an installation)
 */

export type AppMode = 'dashboard' | 'website'

export const APP_MODE = (process.env.APP_MODE || 'website') as AppMode

// Route definitions for each mode
export const MODE_ROUTES = {
  website: [
    '/',
    '/docs',
    '/pricing',
    '/about',
    '/contact',
    '/blog',
    '/access-console',
    '/installations',
    '/auth',
  ],
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
} as const

/**
 * Check if a pathname belongs to the current app mode
 */
export function isRouteAllowedInMode(pathname: string, mode: AppMode): boolean {
  const allowedRoutes = MODE_ROUTES[mode]

  // Safety check: if allowedRoutes is undefined, default to false
  if (!allowedRoutes) {
    console.error(`Invalid app mode: ${mode}. Expected 'dashboard' or 'website'`)
    return false
  }

  // Check if pathname starts with any of the allowed routes
  return allowedRoutes.some((route) => pathname.startsWith(route))
}

/**
 * Get the appropriate redirect URL for a given mode
 */
export function getRedirectForMode(mode: AppMode): string {
  switch (mode) {
    case 'website':
      return '/'
    case 'dashboard':
      return '/dashboard'
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
