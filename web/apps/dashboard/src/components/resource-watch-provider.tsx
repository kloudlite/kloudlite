'use client'

import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useCallback,
  type ReactNode,
} from 'react'
import { useRouter } from 'next/navigation'
import { useSession } from 'next-auth/react'
import { useWebSocket } from '@/lib/hooks/use-websocket'

interface Subscription {
  plural: string
  namespace?: string
  callback: () => void
}

interface ResourceWatchContextValue {
  subscribe: (id: string, sub: Omit<Subscription, 'callback'>, callback: () => void) => void
  unsubscribe: (id: string) => void
}

const ResourceWatchContext = createContext<ResourceWatchContextValue | null>(null)

export function useResourceWatchContext() {
  return useContext(ResourceWatchContext)
}

const DEBOUNCE_MS = 150

export function ResourceWatchProvider({ children }: { children: ReactNode }) {
  const router = useRouter()
  const { status: sessionStatus } = useSession()
  const subsRef = useRef(new Map<string, Subscription>())
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pendingIdsRef = useRef(new Set<string>())
  const connectedOnceRef = useRef(false)
  const routerRef = useRef(router)

  // Keep router ref fresh without re-running effects
  routerRef.current = router

  // Flush: call each pending subscription's callback exactly once
  const flush = useCallback(() => {
    const ids = pendingIdsRef.current
    pendingIdsRef.current = new Set()
    for (const id of ids) {
      const sub = subsRef.current.get(id)
      sub?.callback()
    }
  }, [])

  // Only connect when authenticated
  const wsUrl = sessionStatus === 'authenticated' ? '/api/v1/resource-events-ws' : null

  useWebSocket(wsUrl, {
    onMessage: (data: any) => {
      // Skip control messages
      if (data?.type === 'connected' || data?.type === 'heartbeat') return

      const { plural, namespace } = data

      for (const [id, sub] of subsRef.current.entries()) {
        if (sub.plural === plural && (!sub.namespace || sub.namespace === namespace)) {
          pendingIdsRef.current.add(id)
        }
      }

      if (pendingIdsRef.current.size > 0) {
        if (debounceRef.current) clearTimeout(debounceRef.current)
        debounceRef.current = setTimeout(flush, DEBOUNCE_MS)
      }
    },
    onOpen: () => {
      // On reconnect (not first connect), refresh to catch events we missed
      if (connectedOnceRef.current) {
        routerRef.current.refresh()
      }
      connectedOnceRef.current = true
    },
  })

  // Clean up debounce timer on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [])

  const subscribe = useCallback(
    (id: string, sub: Omit<Subscription, 'callback'>, callback: () => void) => {
      subsRef.current.set(id, { ...sub, callback })
    },
    [],
  )

  const unsubscribe = useCallback((id: string) => {
    subsRef.current.delete(id)
  }, [])

  return (
    <ResourceWatchContext.Provider value={{ subscribe, unsubscribe }}>
      {children}
    </ResourceWatchContext.Provider>
  )
}
