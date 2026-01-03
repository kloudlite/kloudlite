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

  // Rewrites to proxy WebSocket connections to the backend
  async rewrites() {
    return [
      // Environment status WebSocket
      {
        source: '/api/v1/environments/:name/status-ws',
        destination: `${apiUrl}/api/v1/environments/:name/status-ws`,
      },
      // Workspace status WebSocket
      {
        source: '/api/v1/namespaces/:namespace/workspaces/:name/status-ws',
        destination: `${apiUrl}/api/v1/namespaces/:namespace/workspaces/:name/status-ws`,
      },
      // Work machine metrics WebSocket
      {
        source: '/api/v1/work-machines/:name/metrics-ws',
        destination: `${apiUrl}/api/v1/work-machines/:name/metrics-ws`,
      },
      // Service logs WebSocket
      {
        source: '/api/v1/namespaces/:namespace/services/:name/logs-ws',
        destination: `${apiUrl}/api/v1/namespaces/:namespace/services/:name/logs-ws`,
      },
    ]
  },
}

export default nextConfig
