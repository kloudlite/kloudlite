// Environment variable validation and typing
// Validate required environment variables at module load time
function validateEnv() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL
  const webUrl = process.env.NEXT_PUBLIC_WEB_URL
  const appMode = process.env.APP_MODE || process.env.NEXT_PUBLIC_APP_MODE

  // Check if we're in build phase (next build) vs runtime
  // During build, Next.js sets NODE_ENV=production but we don't have runtime env vars
  const isBuildTime =
    typeof window === 'undefined' && process.env.NEXT_PHASE === 'phase-production-build'

  // Registration mode doesn't need API_URL/WEB_URL as it uses Supabase directly
  const isRegistrationMode = appMode === 'registration'

  // In production runtime (not build), fail fast if critical env vars are missing
  // Skip validation for registration mode as it doesn't use the backend API
  if (process.env.NODE_ENV === 'production' && !isBuildTime && !isRegistrationMode && !apiUrl) {
    throw new Error(
      'CRITICAL: NEXT_PUBLIC_API_URL environment variable is not set. ' +
        'The application cannot function without this configuration. ' +
        'Please set NEXT_PUBLIC_API_URL in your environment variables.',
    )
  }

  if (process.env.NODE_ENV === 'production' && !isBuildTime && !isRegistrationMode && !webUrl) {
    throw new Error(
      'CRITICAL: NEXT_PUBLIC_WEB_URL environment variable is not set. ' +
        'The application cannot function without this configuration. ' +
        'Please set NEXT_PUBLIC_WEB_URL in your environment variables.',
    )
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
