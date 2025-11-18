// Environment variable validation and typing
// Validate required environment variables at module load time
function validateEnv() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL
  const appMode = process.env.APP_MODE || process.env.NEXT_PUBLIC_APP_MODE

  // Check if we're in build phase (next build) vs runtime
  // During build, Next.js sets NODE_ENV=production but we don't have runtime env vars
  const isBuildTime =
    typeof window === 'undefined' && process.env.NEXT_PHASE === 'phase-production-build'

  // Website and console modes use Supabase for managing installations (don't need API_URL)
  // Dashboard mode runs in tenant installations and uses the API server (needs API_URL)
  const isSupabaseMode = appMode === 'website' || appMode === 'console'
  const isDashboardMode = appMode === 'dashboard'

  // Dashboard mode requires API_URL
  if (process.env.NODE_ENV === 'production' && !isBuildTime && isDashboardMode && !apiUrl) {
    throw new Error(
      'CRITICAL: NEXT_PUBLIC_API_URL environment variable is not set. ' +
        'Dashboard mode requires this to connect to the API server. ' +
        'Please set NEXT_PUBLIC_API_URL in your environment variables.',
    )
  }

  // In development or build time, warn but allow localhost fallback
  if ((process.env.NODE_ENV === 'development' || isBuildTime) && !apiUrl) {
    if (!isBuildTime) {
      console.warn('⚠️  NEXT_PUBLIC_API_URL is not set. Falling back to http://localhost:8080')
    }
  }

  return {
    apiUrl: apiUrl || 'http://localhost:8080',
    env: process.env.NEXT_PUBLIC_ENV || 'development',
    isDevelopment: process.env.NODE_ENV === 'development',
    isProduction: process.env.NODE_ENV === 'production',
  }
}

// Validate and export environment configuration
export const env = validateEnv()

// Type for environment configuration
export type Environment = typeof env
