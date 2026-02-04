'use client'

import { useEnvironmentStatus } from '@/lib/hooks/use-environment-status'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

interface EnvironmentStatusIndicatorProps {
  environmentName: string
  initialState?: string
  className?: string
  showLoader?: boolean
}

function getStatusStyles(state: string | undefined) {
  switch (state) {
    case 'active':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
    case 'inactive':
      return 'bg-secondary text-secondary-foreground'
    case 'activating':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    case 'deactivating':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
    case 'deleting':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    case 'error':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-secondary text-secondary-foreground'
  }
}

function isTransitionalState(state: string | undefined) {
  return state === 'activating' || state === 'deactivating' || state === 'deleting'
}

export function EnvironmentStatusIndicator({
  environmentName,
  initialState,
  className,
  showLoader = true,
}: EnvironmentStatusIndicatorProps) {
  const router = useRouter()
  const { state, isConnected } = useEnvironmentStatus(environmentName, {
    enabled: true,
    onStateChange: (newState) => {
      // Refresh the page when transitioning to a stable state
      if (!isTransitionalState(newState)) {
        router.refresh()
      }
    },
  })

  // Use SSE state if connected, otherwise fall back to initial state
  const displayState = isConnected && state ? state : initialState

  // Refresh page data when status changes to ensure consistency
  useEffect(() => {
    if (isConnected && state && state !== initialState) {
      router.refresh()
    }
  }, [isConnected, state, initialState, router])

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-2.5 py-0.5 text-xs font-medium',
        getStatusStyles(displayState),
        className
      )}
    >
      {showLoader && isTransitionalState(displayState) && (
        <Loader2 className="h-3 w-3 animate-spin" />
      )}
      {displayState || 'unknown'}
    </span>
  )
}
