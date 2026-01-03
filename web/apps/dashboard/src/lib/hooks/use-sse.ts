'use client'

import { useEffect, useRef, useCallback, useState } from 'react'

// Reconnection constants
const MAX_RECONNECT_ATTEMPTS = 5
const BASE_RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 10000

// Fetch auth token for SSE connections
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

export interface UseSSEOptions {
  /** Whether the SSE connection is enabled */
  enabled?: boolean
  /** Called when a message is received */
  onMessage?: (data: string) => void
  /** Called when connection is established */
  onOpen?: () => void
  /** Called when connection is closed or fails permanently */
  onError?: (error: string) => void
}

export interface UseSSEResult {
  isConnected: boolean
  error: string | null
  reconnect: () => void
}

export function useSSE(
  url: string | null | undefined,
  options: UseSSEOptions = {}
): UseSSEResult {
  const { enabled = true, onMessage, onOpen, onError } = options

  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCleaningUpRef = useRef(false)
  const tokenRef = useRef<string | null>(null)

  // Store callbacks in refs to avoid recreating connect function
  const onMessageRef = useRef(onMessage)
  const onOpenRef = useRef(onOpen)
  const onErrorRef = useRef(onError)

  // Update refs when callbacks change
  useEffect(() => {
    onMessageRef.current = onMessage
    onOpenRef.current = onOpen
    onErrorRef.current = onError
  }, [onMessage, onOpen, onError])

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

  const connect = useCallback(async () => {
    if (!url || !enabled || isCleaningUpRef.current) return

    // Close existing connection if any
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Get auth token if not cached
    if (!tokenRef.current) {
      tokenRef.current = await getAuthToken()
    }

    // Build URL with auth token
    let sseUrl = url
    if (tokenRef.current) {
      const separator = sseUrl.includes('?') ? '&' : '?'
      sseUrl = `${sseUrl}${separator}token=${encodeURIComponent(tokenRef.current)}`
    }

    const eventSource = new EventSource(sseUrl, { withCredentials: true })
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setError(null)
      reconnectAttemptsRef.current = 0
      onOpenRef.current?.()
    }

    eventSource.onmessage = (event) => {
      onMessageRef.current?.(event.data)
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
