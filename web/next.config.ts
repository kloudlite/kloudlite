import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // Enable standalone output for Docker deployment
  output: 'standalone',

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
