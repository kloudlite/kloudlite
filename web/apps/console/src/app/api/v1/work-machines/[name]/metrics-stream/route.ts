import { getAuthToken } from '@/lib/get-session'
import { env } from '@kloudlite/lib'
import { NextRequest } from 'next/server'

export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

export async function GET(request: NextRequest, { params }: { params: Promise<{ name: string }> }) {
  const { name } = await params

  try {
    // Get authentication token
    const token = await getAuthToken()
    if (!token) {
      return new Response('Unauthorized', { status: 401 })
    }

    // Create connection to backend SSE endpoint
    const backendUrl = `${env.apiUrl}/api/v1/work-machines/${encodeURIComponent(name)}/metrics-stream`

    const backendResponse = await fetch(backendUrl, {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'text/event-stream',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
      },
      signal: request.signal, // Forward abort signal from client
    })

    if (!backendResponse.ok) {
      const errorText = await backendResponse.text()
      console.error(
        'Backend SSE connection failed:',
        backendResponse.status,
        errorText,
      )
      return new Response(`Backend error: ${backendResponse.statusText}`, {
        status: backendResponse.status,
      })
    }

    // Check if the backend actually returned SSE
    const contentType = backendResponse.headers.get('content-type')
    if (!contentType?.includes('text/event-stream')) {
      console.error('Backend did not return SSE stream:', contentType)
      return new Response('Backend did not return event stream', { status: 500 })
    }

    // Create a TransformStream to proxy SSE events
    const { readable, writable } = new TransformStream()
    const writer = writable.getWriter()

    // Stream the response from backend to client
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
            console.log('Backend SSE stream ended')
            break
          }

          // Forward the chunk to the client
          await writer.write(value)
        }
      } catch (error) {
        console.error('Error proxying SSE stream:', error)
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
        'X-Accel-Buffering': 'no', // Disable nginx buffering
      },
    })
  } catch (error) {
    console.error('SSE proxy error:', error)
    const err = error instanceof Error ? error : new Error('Unknown error')
    return new Response(`Error: ${err.message}`, { status: 500 })
  }
}
