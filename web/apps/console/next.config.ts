import type { NextConfig } from 'next'
import path from 'path'

const nextConfig: NextConfig = {
  // Enable standalone output for Docker deployment
  output: 'standalone',

  // Set the workspace root for file tracing
  outputFileTracingRoot: path.join(__dirname, '../../'),

  // Compress responses
  compress: true,

  // Security headers are now set dynamically in middleware.ts
  // to support per-request subdomain-based CSP policies

  // Enable experimental features
  experimental: {
    // Optimize package imports
    optimizePackageImports: ['@radix-ui/react-icons', '@radix-ui/react-dialog'],
  },
}

export default nextConfig
