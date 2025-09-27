// Environment variable validation and typing
const requiredEnvVars = ['NEXT_PUBLIC_API_URL'] as const

// Validate environment variables
for (const envVar of requiredEnvVars) {
  if (!process.env[envVar]) {
    throw new Error(`Missing required environment variable: ${envVar}`)
  }
}

export const env = {
  apiUrl: process.env.NEXT_PUBLIC_API_URL!,
  env: process.env.NEXT_PUBLIC_ENV || 'development',
  isDevelopment: process.env.NODE_ENV === 'development',
  isProduction: process.env.NODE_ENV === 'production',
} as const

// Type for environment configuration
export type Environment = typeof env