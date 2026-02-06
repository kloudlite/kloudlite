'use client'

import { useEffect, useId } from 'react'
import { useRouter } from 'next/navigation'
import { useResourceWatchContext } from '@/components/resource-watch-provider'

/**
 * Subscribe to real-time resource changes via WebSocket.
 * When a matching change arrives, calls router.refresh().
 *
 * @param plural - resource type, e.g. 'workspaces', 'environments', 'workmachines'
 * @param namespace - optional namespace filter; omit for cluster-scoped or all namespaces
 */
export function useResourceWatch(plural: string, namespace?: string) {
  const ctx = useResourceWatchContext()
  const router = useRouter()
  const id = useId()

  useEffect(() => {
    if (!ctx) return

    ctx.subscribe(id, { plural, namespace }, () => {
      router.refresh()
    })

    return () => {
      ctx.unsubscribe(id)
    }
  }, [ctx, id, plural, namespace, router])
}
