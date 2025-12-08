import { getAuthToken } from '@/lib/get-session'
import { env } from '@/lib/env'
import { NextRequest } from 'next/server'

export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ namespace: string; name: string }> }
) {
  const { namespace, name } = await params

  try {
    // Get JWT token for backend API authentication
    const token = await getAuthToken()
    if (!token) {
      return new Response('Unauthorized', { status: 401 })
    }

    // Get query parameters
    const searchParams = request.nextUrl.searchParams
    const follow = searchParams.get('follow') || 'true'
    const tailLines = searchParams.get('tailLines') || '200'
    const timestamps = searchParams.get('timestamps') || 'false'
    const container = searchParams.get('container') || ''

    // Build backend URL with query params
    const queryParams = new URLSearchParams({
      follow,
      tailLines,
      timestamps,
    })
    if (container) {
      queryParams.set('container', container)
    }

    const backendUrl = `${env.apiUrl}/api/v1/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(name)}/logs?${queryParams}`

    const backendResponse = await fetch(backendUrl, {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'text/event-stream',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
      },
      signal: request.signal,
    })

    if (!backendResponse.ok) {
      const errorText = await backendResponse.text()
      console.error(
        'Backend logs stream failed:',
        backendResponse.status,
        errorText
      )
      return new Response(
        JSON.stringify({ error: errorText || backendResponse.statusText }),
        {
          status: backendResponse.status,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }

    // Check if the backend returned SSE
    const contentType = backendResponse.headers.get('content-type')
    if (!contentType?.includes('text/event-stream')) {
      console.error('Backend did not return SSE stream:', contentType)
      return new Response('Backend did not return event stream', { status: 500 })
    }

    // Create a TransformStream to proxy SSE events
    const { readable, writable } = new TransformStream()
    const writer = writable.getWriter()

    if (!backendResponse.body) {
      return new Response('No body in backend response', { status: 500 })
    }

    const reader = backendResponse.body.getReader()

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
        console.error('Error proxying logs stream:', error)
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
        Connection: 'keep-alive',
        'X-Accel-Buffering': 'no',
      },
    })
  } catch (error) {
    console.error('Logs stream proxy error:', error)
    const err = error instanceof Error ? error : new Error('Unknown error')
    return new Response(JSON.stringify({ error: err.message }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }
}
