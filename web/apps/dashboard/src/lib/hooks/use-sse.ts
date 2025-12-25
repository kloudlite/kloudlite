'use client'

import { useEffect, useRef, useCallback, useState } from 'react'

// Reconnection constants
const MAX_RECONNECT_ATTEMPTS = 10
const BASE_RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 30000

export interface UseSSEOptions<T> {
  /** Whether the SSE connection is enabled */
  enabled?: boolean
  /** Called when a message is received (for default 'message' events) */
  onMessage?: (data: string) => void
  /** Called when connection is established */
  onOpen?: () => void
  /** Called when connection is closed or fails permanently */
  onError?: (error: string) => void
  /** Event handlers for named events (e.g., 'status', 'metrics') */
  eventHandlers?: Record<string, (data: T) => void>
  /** Parse JSON automatically for event handlers */
  parseJson?: boolean
}

export interface UseSSEResult {
  isConnected: boolean
  error: string | null
  reconnect: () => void
}

export function useSSE<T = unknown>(
  url: string | null | undefined,
  options: UseSSEOptions<T> = {}
): UseSSEResult {
  const {
    enabled = true,
    onMessage,
    onOpen,
    onError,
    eventHandlers,
    parseJson = true,
  } = options

  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCleaningUpRef = useRef(false)

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
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
  }, [])

  const connect = useCallback(() => {
    if (!url || !enabled || isCleaningUpRef.current) return

    // Close existing connection if any
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setError(null)
      reconnectAttemptsRef.current = 0
      onOpenRef.current?.()
    }

    // Handle default message events
    if (onMessageRef.current) {
      eventSource.onmessage = (event) => {
        onMessageRef.current?.(event.data)
      }
    }

    // Handle named events
    if (eventHandlersRef.current) {
      Object.entries(eventHandlersRef.current).forEach(([eventName, handler]) => {
        eventSource.addEventListener(eventName, (event: MessageEvent) => {
          try {
            const data = parseJson ? JSON.parse(event.data) : event.data
            handler(data)
          } catch (err) {
            console.error(`Failed to parse SSE event '${eventName}':`, err)
          }
        })
      })
    }

    eventSource.onerror = () => {
      setIsConnected(false)
      eventSource.close()
      eventSourceRef.current = null

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
          `SSE connection lost, reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS})`
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
  }, [url, enabled, parseJson])

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
