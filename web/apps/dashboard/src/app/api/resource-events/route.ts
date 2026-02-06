import { resourceStore, type ResourceChangeEvent } from '@/lib/resource-store'
import { getSession } from '@/lib/get-session'

export const dynamic = 'force-dynamic'

export async function GET() {
  const session = await getSession()
  if (!session) {
    return new Response('Unauthorized', { status: 401 })
  }

  const encoder = new TextEncoder()
  let cleaned = false
  let cleanup: (() => void) | undefined

  const stream = new ReadableStream({
    start(controller) {
      const send = (data: string) => {
        if (cleaned) return
        try {
          controller.enqueue(encoder.encode(`data: ${data}\n\n`))
        } catch {
          doCleanup()
        }
      }

      const onChange = (event: ResourceChangeEvent) => {
        send(JSON.stringify(event))
      }

      const doCleanup = () => {
        if (cleaned) return
        cleaned = true
        resourceStore.emitter.off('change', onChange)
        clearInterval(heartbeat)
      }

      // Expose to cancel()
      cleanup = doCleanup

      send(JSON.stringify({ type: 'connected' }))

      resourceStore.emitter.on('change', onChange)

      const heartbeat = setInterval(() => {
        send(JSON.stringify({ type: 'heartbeat' }))
      }, 30000)
    },
    cancel() {
      cleanup?.()
    },
  })

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache, no-transform',
      Connection: 'keep-alive',
      'X-Accel-Buffering': 'no',
    },
  })
}
