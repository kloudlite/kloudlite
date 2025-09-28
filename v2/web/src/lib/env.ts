// Environment variable validation and typing
const requiredEnvVars = ['NEXT_PUBLIC_API_URL'] as const

// Note: In Next.js, process.env values are replaced at build time
// Client-side validation of process.env doesn't work as expected

export const env = {
  apiUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  env: process.env.NEXT_PUBLIC_ENV || 'development',
  isDevelopment: process.env.NODE_ENV === 'development',
  isProduction: process.env.NODE_ENV === 'production',
} as const

// Type for environment configuration
export type Environment = typeof env