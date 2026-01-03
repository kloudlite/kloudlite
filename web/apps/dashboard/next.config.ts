import type { NextConfig } from 'next'
import path from 'path'

const apiUrl = process.env.API_URL || 'http://api-server.kloudlite.svc.cluster.local'

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

  // Rewrites to proxy API requests to the backend
  // SSE (Server-Sent Events) works through rewrites since it's regular HTTP
  async rewrites() {
    return [
      // Service logs SSE
      {
        source: '/api/v1/namespaces/:namespace/services/:name/logs',
        destination: `${apiUrl}/api/v1/namespaces/:namespace/services/:name/logs`,
      },
      // Work machine metrics SSE
      {
        source: '/api/v1/work-machines/:name/metrics',
        destination: `${apiUrl}/api/v1/work-machines/:name/metrics`,
      },
      // Environment status
      {
        source: '/api/v1/environments/:name/status',
        destination: `${apiUrl}/api/v1/environments/:name/status`,
      },
    ]
  },
}

export default nextConfig
