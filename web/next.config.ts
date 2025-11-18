import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // Enable standalone output for Docker deployment
  output: 'standalone',

  // Compress responses
  compress: true,

  // Security headers to block third-party tracking scripts
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'Content-Security-Policy',
            value: "script-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' http://localhost:* https://localhost:* ws://localhost:* wss://localhost:*;",
          },
        ],
      },
    ]
  },

  // Enable experimental features
  experimental: {
    // Optimize package imports
    optimizePackageImports: ['@radix-ui/react-icons', '@radix-ui/react-dialog'],
  },
}

export default nextConfig
