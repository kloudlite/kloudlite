// Environment variable validation and typing
const requiredEnvVars = ['NEXT_PUBLIC_API_URL'] as const

// Validate required environment variables at module load time
function validateEnv() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL
  const webUrl = process.env.NEXT_PUBLIC_WEB_URL

  // In production, fail fast if critical env vars are missing
  if (process.env.NODE_ENV === 'production' && !apiUrl) {
    throw new Error(
      'CRITICAL: NEXT_PUBLIC_API_URL environment variable is not set. ' +
      'The application cannot function without this configuration. ' +
      'Please set NEXT_PUBLIC_API_URL in your environment variables.'
    )
  }

  if (process.env.NODE_ENV === 'production' && !webUrl) {
    throw new Error(
      'CRITICAL: NEXT_PUBLIC_WEB_URL environment variable is not set. ' +
      'The application cannot function without this configuration. ' +
      'Please set NEXT_PUBLIC_WEB_URL in your environment variables.'
    )
  }

  // In development, warn but allow localhost fallback
  if (process.env.NODE_ENV === 'development' && !apiUrl) {
    console.warn(
      '⚠️  NEXT_PUBLIC_API_URL is not set. Falling back to http://localhost:8080'
    )
  }

  if (process.env.NODE_ENV === 'development' && !webUrl) {
    console.warn(
      '⚠️  NEXT_PUBLIC_WEB_URL is not set. Falling back to http://localhost:3000'
    )
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