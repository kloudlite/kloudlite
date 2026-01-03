// Environment variable validation and typing
// Console app uses Supabase for managing tenant workspace installations
function validateEnv() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL
  const webUrl = process.env.NEXT_PUBLIC_WEB_URL

  // Check if we're in build phase (next build) vs runtime
  const isBuildTime =
    typeof window === 'undefined' && process.env.NEXT_PHASE === 'phase-production-build'

  // NEXT_PUBLIC_API_URL is optional - WebSocket connections are proxied through the server
  if (process.env.NODE_ENV === 'production' && !isBuildTime && !apiUrl) {
    console.warn('NEXT_PUBLIC_API_URL not set, using relative URLs')
  }

  // In development or build time, warn but allow localhost fallback
  if ((process.env.NODE_ENV === 'development' || isBuildTime) && !apiUrl) {
    if (!isBuildTime) {
      console.warn('⚠️  NEXT_PUBLIC_API_URL is not set. Falling back to http://localhost:8080')
    }
  }

  if ((process.env.NODE_ENV === 'development' || isBuildTime) && !webUrl) {
    if (!isBuildTime) {
      console.warn('⚠️  NEXT_PUBLIC_WEB_URL is not set. Falling back to http://localhost:3000')
    }
  }

  return {
    apiUrl: apiUrl || 'http://localhost:8080',
    webUrl: webUrl || 'http://localhost:3000',
    env: process.env.NEXT_PUBLIC_ENV || 'development',
    isDevelopment: process.env.NODE_ENV === 'development',
    isProduction: process.env.NODE_ENV === 'production',
  }
}

// Validate and export environment configuration
export const env = validateEnv()

// Type for environment configuration
export type Environment = typeof env
