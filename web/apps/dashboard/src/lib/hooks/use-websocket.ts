'use client'

import { useEffect, useRef, useCallback, useState } from 'react'

// Reconnection constants
const MAX_RECONNECT_ATTEMPTS = 10
const BASE_RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 30000

// Convert relative path to WebSocket URL using current origin
// WebSocket connections go through the custom server proxy
function toWebSocketUrl(path: string): string {
  // If path is already a full WebSocket URL, use it directly
  if (path.startsWith('ws://') || path.startsWith('wss://')) {
    return path
  }

  // Use current browser origin and convert to WebSocket protocol
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host

  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`

  return `${protocol}//${host}${normalizedPath}`
}

export interface UseWebSocketOptions<T> {
  /** Whether the WebSocket connection is enabled */
  enabled?: boolean
  /** Called when a message is received */
  onMessage?: (data: T) => void
  /** Called when connection is established */
  onOpen?: () => void
  /** Called when connection is closed or fails permanently */
  onError?: (error: string) => void
  /** Event handlers for typed events (based on event.type field) */
  eventHandlers?: Record<string, (data: T) => void>
}

export interface UseWebSocketResult {
  isConnected: boolean
  error: string | null
  reconnect: () => void
}

// Fetch auth token for WebSocket connections
async function getAuthToken(): Promise<string | null> {
  try {
    const response = await fetch('/api/auth/token')
    if (!response.ok) return null
    const data = await response.json()
    return data.token
  } catch {
    return null
  }
}

export function useWebSocket<T = unknown>(
  url: string | null | undefined,
  options: UseWebSocketOptions<T> = {}
): UseWebSocketResult {
  const {
    enabled = true,
    onMessage,
    onOpen,
    onError,
    eventHandlers,
  } = options

  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCleaningUpRef = useRef(false)
  const tokenRef = useRef<string | null>(null)

  // Store callbacks in refs to avoid recreating connect function
  const onMessageRef = useRef(onMessage)
  const onOpenRef = useRef(onOpen)
  const onErrorRef = useRef(onError)
  const eventHandlersRef = useRef(eventHandlers)

  // Update refs when callbacks change
  useEffect(() => {
    onMessageRef.current = onMessage
    onOpenRef.current = onOpen
    onErrorRef.current = onError
    eventHandlersRef.current = eventHandlers
  }, [onMessage, onOpen, onError, eventHandlers])

  const cleanup = useCallback(() => {
    isCleaningUpRef.current = true
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }, [])

  const connect = useCallback(async () => {
    if (!url || !enabled || isCleaningUpRef.current) return

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close()
    }

    // Get auth token if not cached
    if (!tokenRef.current) {
      tokenRef.current = await getAuthToken()
    }

    // Build WebSocket URL - use direct connection to API server
    let wsUrl = toWebSocketUrl(url)
    if (tokenRef.current) {
      const separator = wsUrl.includes('?') ? '&' : '?'
      wsUrl = `${wsUrl}${separator}token=${encodeURIComponent(tokenRef.current)}`
    }

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setIsConnected(true)
      setError(null)
      reconnectAttemptsRef.current = 0
      onOpenRef.current?.()
    }

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as { type?: string; data?: T } | T

        // Check if message is in {type, data} format (backend standard)
        if (typeof message === 'object' && message !== null && 'type' in message) {
          const typedMessage = message as { type: string; data?: T }
          if (eventHandlersRef.current?.[typedMessage.type]) {
            // Pass the nested data to the handler, or the whole message if no data field
            const payload = typedMessage.data !== undefined ? typedMessage.data : (typedMessage as unknown as T)
            eventHandlersRef.current[typedMessage.type](payload)
          } else if (onMessageRef.current) {
            onMessageRef.current(typedMessage.data !== undefined ? typedMessage.data : (typedMessage as unknown as T))
          }
        } else if (onMessageRef.current) {
          onMessageRef.current(message as T)
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }

    ws.onerror = () => {
      setIsConnected(false)
    }

    ws.onclose = () => {
      setIsConnected(false)
      wsRef.current = null

      // Don't reconnect if cleaning up
      if (isCleaningUpRef.current) return

      // Attempt reconnection with exponential backoff
      if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
        const delay = Math.min(
          BASE_RECONNECT_DELAY * Math.pow(2, reconnectAttemptsRef.current),
          MAX_RECONNECT_DELAY
        )
        reconnectAttemptsRef.current++

        console.log(
          `WebSocket connection lost, reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS})`
        )

        reconnectTimeoutRef.current = setTimeout(() => {
          if (!isCleaningUpRef.current) {
            connect()
          }
        }, delay)
      } else {
        const errorMsg = 'Connection failed after multiple attempts'
        setError(errorMsg)
        onErrorRef.current?.(errorMsg)
      }
    }
  }, [url, enabled])

  useEffect(() => {
    if (!enabled || !url) {
      setIsConnected(false)
      return
    }

    isCleaningUpRef.current = false
    reconnectAttemptsRef.current = 0
    connect()

    return () => {
      cleanup()
    }
  }, [url, enabled, connect, cleanup])

  const reconnect = useCallback(() => {
    cleanup()
    isCleaningUpRef.current = false
    reconnectAttemptsRef.current = 0
    setError(null)
    connect()
  }, [connect, cleanup])

  return {
    isConnected,
    error,
    reconnect,
  }
}
