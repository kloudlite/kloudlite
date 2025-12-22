import { NextRequest } from 'next/server'
import { getInstallationBySecretKey } from '@/lib/console/supabase-storage-service'

export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

const ANTHROPIC_API_URL = 'https://api.anthropic.com/v1/messages'

/**
 * Claude API Proxy
 * Forwards requests to Anthropic's Claude API
 * Requires valid installation secret key for authentication
 * Supports both streaming (SSE) and non-streaming responses
 *
 * Authentication:
 * - Claude Code sends the secret key via x-api-key header
 * - The secret key is obtained during installation via verify-key endpoint
 */
export async function POST(request: NextRequest) {
  // Get secret key from x-api-key header (Claude Code compatible)
  const secretKey = request.headers.get('x-api-key')

  if (!secretKey) {
    return Response.json(
      { error: { type: 'authentication_error', message: 'Missing API key' } },
      { status: 401 }
    )
  }

  // Validate secret key against database
  const installation = await getInstallationBySecretKey(secretKey)

  if (!installation) {
    return Response.json(
      { error: { type: 'authentication_error', message: 'Invalid API key' } },
      { status: 401 }
    )
  }

  console.log('[Claude Proxy] Authenticated installation:', installation.id, installation.subdomain)

  const apiKey = process.env.ANTHROPIC_API_KEY
  if (!apiKey) {
    console.error('[Claude Proxy] ANTHROPIC_API_KEY environment variable not set')
    return Response.json(
      { error: { type: 'configuration_error', message: 'API key not configured' } },
      { status: 500 }
    )
  }

  try {
    const body = await request.json()
    const isStreaming = body.stream === true

    // Forward request to Anthropic API
    const response = await fetch(ANTHROPIC_API_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify(body),
    })

    // Handle error responses
    if (!response.ok) {
      const errorBody = await response.text()
      console.error('[Claude Proxy] Anthropic API error:', response.status, errorBody)

      try {
        const errorJson = JSON.parse(errorBody)
        return Response.json(errorJson, { status: response.status })
      } catch {
        return Response.json(
          { error: { type: 'api_error', message: errorBody } },
          { status: response.status }
        )
      }
    }

    // Handle streaming response
    if (isStreaming) {
      if (!response.body) {
        return Response.json(
          { error: { type: 'stream_error', message: 'No response body from Anthropic API' } },
          { status: 500 }
        )
      }

      // Create a TransformStream to proxy SSE events
      const { readable, writable } = new TransformStream()
      const writer = writable.getWriter()
      const reader = response.body.getReader()

      // Pipe the stream in the background
      ;(async () => {
        try {
          while (true) {
            const { done, value } = await reader.read()
            if (done) {
              break
            }
            await writer.write(value)
          }
        } catch (error) {
          console.error('[Claude Proxy] Error proxying SSE stream:', error)
        } finally {
          try {
            await writer.close()
          } catch {
            // Ignore close errors
          }
          try {
            reader.releaseLock()
          } catch {
            // Ignore release errors
          }
        }
      })()

      // Return SSE response with proper headers
      return new Response(readable, {
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache, no-transform',
          'Connection': 'keep-alive',
          'X-Accel-Buffering': 'no', // Disable nginx buffering
        },
      })
    }

    // Handle non-streaming response
    const responseData = await response.json()
    return Response.json(responseData)
  } catch (error) {
    console.error('[Claude Proxy] Request error:', error)
    const message = error instanceof Error ? error.message : 'Unknown error'
    return Response.json(
      { error: { type: 'proxy_error', message } },
      { status: 500 }
    )
  }
}
