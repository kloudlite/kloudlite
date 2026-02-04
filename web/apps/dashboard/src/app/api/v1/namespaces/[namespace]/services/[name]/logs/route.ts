import { getSession } from '@/lib/get-session'
import { env } from '@/lib/env'
import { NextRequest } from 'next/server'

export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

interface K8sService {
  spec: {
    selector?: Record<string, string>
  }
}

interface K8sPodList {
  items: Array<{
    metadata: {
      name: string
      namespace: string
    }
    spec: {
      containers: Array<{
        name: string
      }>
    }
    status: {
      phase: string
    }
  }>
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ namespace: string; name: string }> }
) {
  const { namespace, name } = await params

  try {
    // Check user is authenticated
    const session = await getSession()
    if (!session?.user) {
      return new Response('Unauthorized', { status: 401 })
    }

    // Get query parameters
    const searchParams = request.nextUrl.searchParams
    const follow = searchParams.get('follow') === 'true'
    const tailLines = searchParams.get('tailLines') || '200'
    const timestamps = searchParams.get('timestamps') === 'true'
    const container = searchParams.get('container') || ''

    // Step 1: Get the service to find its selector
    // Note: kubectl proxy handles authentication, so no Authorization header needed
    const serviceUrl = `${env.apiUrl}/api/v1/namespaces/${encodeURIComponent(namespace)}/services/${encodeURIComponent(name)}`
    const serviceResponse = await fetch(serviceUrl, {
      headers: {
        Accept: 'application/json',
      },
    })

    if (!serviceResponse.ok) {
      const errorText = await serviceResponse.text()
      console.error('Failed to get service:', serviceResponse.status, errorText)
      return new Response(
        JSON.stringify({ error: `Service not found: ${errorText}` }),
        {
          status: serviceResponse.status,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }

    const service: K8sService = await serviceResponse.json()
    const selector = service.spec?.selector

    if (!selector || Object.keys(selector).length === 0) {
      return new Response(
        JSON.stringify({ error: 'Service has no selector, cannot find pods' }),
        {
          status: 400,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }

    // Step 2: List pods matching the service selector
    const labelSelector = Object.entries(selector)
      .map(([k, v]) => `${k}=${v}`)
      .join(',')

    const podsUrl = `${env.apiUrl}/api/v1/namespaces/${encodeURIComponent(namespace)}/pods?labelSelector=${encodeURIComponent(labelSelector)}`
    const podsResponse = await fetch(podsUrl, {
      headers: {
        Accept: 'application/json',
      },
    })

    if (!podsResponse.ok) {
      const errorText = await podsResponse.text()
      console.error('Failed to list pods:', podsResponse.status, errorText)
      return new Response(
        JSON.stringify({ error: `Failed to list pods: ${errorText}` }),
        {
          status: podsResponse.status,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }

    const podList: K8sPodList = await podsResponse.json()

    // Filter to running pods only
    const runningPods = podList.items.filter(
      (pod) => pod.status.phase === 'Running'
    )

    if (runningPods.length === 0) {
      return new Response(
        JSON.stringify({ error: 'No running pods found for this service' }),
        {
          status: 404,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }

    // Step 3: Stream logs from the first running pod
    // TODO: Consider aggregating logs from all pods
    const pod = runningPods[0]
    const containerName =
      container || pod.spec.containers[0]?.name || ''

    // Build pod logs URL
    const queryParams = new URLSearchParams()
    if (follow) queryParams.set('follow', 'true')
    if (tailLines) queryParams.set('tailLines', tailLines)
    if (timestamps) queryParams.set('timestamps', 'true')
    if (containerName) queryParams.set('container', containerName)

    const podLogsUrl = `${env.apiUrl}/api/v1/namespaces/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(pod.metadata.name)}/log?${queryParams}`

    const logsResponse = await fetch(podLogsUrl, {
      signal: request.signal,
    })

    if (!logsResponse.ok) {
      const errorText = await logsResponse.text()
      console.error('Failed to get pod logs:', logsResponse.status, errorText)
      return new Response(
        JSON.stringify({ error: `Failed to get logs: ${errorText}` }),
        {
          status: logsResponse.status,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }

    // Check if following logs (streaming) or one-shot
    if (!follow) {
      // One-shot: return logs as SSE format for consistency
      const logsText = await logsResponse.text()
      const lines = logsText.split('\n').filter((line) => line.trim())

      // Convert to SSE format
      const sseData = lines.map((line) => `data: ${line}\n\n`).join('')

      return new Response(sseData, {
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache, no-transform',
          Connection: 'keep-alive',
          'X-Accel-Buffering': 'no',
        },
      })
    }

    // Streaming: Create a TransformStream to convert plain text logs to SSE
    const { readable, writable } = new TransformStream()
    const writer = writable.getWriter()
    const encoder = new TextEncoder()

    if (!logsResponse.body) {
      return new Response('No body in logs response', { status: 500 })
    }

    const reader = logsResponse.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    // Pipe the stream in the background, converting to SSE format
    ;(async () => {
      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) {
            // Flush any remaining buffer
            if (buffer.trim()) {
              await writer.write(encoder.encode(`data: ${buffer}\n\n`))
            }
            break
          }

          // Decode the chunk and split into lines
          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')

          // Keep the last incomplete line in the buffer
          buffer = lines.pop() || ''

          // Send complete lines as SSE events
          for (const line of lines) {
            if (line.trim()) {
              await writer.write(encoder.encode(`data: ${line}\n\n`))
            }
          }
        }
      } catch (error) {
        console.error('Error streaming logs:', error)
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

    // Return SSE response
    return new Response(readable, {
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache, no-transform',
        Connection: 'keep-alive',
        'X-Accel-Buffering': 'no',
      },
    })
  } catch (error) {
    console.error('Logs stream error:', error)
    const err = error instanceof Error ? error : new Error('Unknown error')
    return new Response(JSON.stringify({ error: err.message }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    })
  }
}
