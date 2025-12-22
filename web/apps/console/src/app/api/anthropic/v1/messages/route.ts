import { NextRequest } from 'next/server'
import { jwtVerify } from 'jose'

export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

const ANTHROPIC_API_URL = 'https://api.anthropic.com/v1/messages'

/**
 * Validates JWT token from Authorization header
 * Returns user payload if valid, null if invalid
 */
async function validateToken(request: NextRequest): Promise<{ username?: string; email?: string } | null> {
  const authHeader = request.headers.get('Authorization')
  if (!authHeader?.startsWith('Bearer ')) {
    return null
  }

  const token = authHeader.slice(7)
  const jwtSecret = process.env.JWT_SECRET

  if (!jwtSecret) {
    console.error('[Claude Proxy] JWT_SECRET environment variable not set')
    return null
  }

  try {
    const secret = new TextEncoder().encode(jwtSecret)
    const { payload } = await jwtVerify(token, secret)
    return {
      username: payload.username as string | undefined,
      email: payload.email as string | undefined,
    }
  } catch (error) {
    console.error('[Claude Proxy] JWT verification failed:', error)
    return null
  }
}

/**
 * Claude API Proxy
 * Forwards requests to Anthropic's Claude API
 * Requires valid Kloudlite JWT token for authentication
 * Supports both streaming (SSE) and non-streaming responses
 */
export async function POST(request: NextRequest) {
  // Validate JWT token
  const user = await validateToken(request)
  if (!user) {
    return Response.json(
      { error: { type: 'authentication_error', message: 'Invalid or missing authorization token' } },
      { status: 401 }
    )
  }

  console.log('[Claude Proxy] Authenticated user:', user.username || user.email)

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
