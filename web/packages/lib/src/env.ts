// Environment variable validation and typing
// Validate required environment variables at module load time
function validateEnv() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL

  // Check if we're in build phase (next build) vs runtime
  // During build, Next.js sets NODE_ENV=production but we don't have runtime env vars
  const isBuildTime =
    typeof window === 'undefined' && process.env.NEXT_PHASE === 'phase-production-build'

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
