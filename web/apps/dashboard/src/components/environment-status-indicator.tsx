'use client'

import { useEnvironmentStatus } from '@/lib/hooks/use-environment-status'
import { Loader2 } from 'lucide-react'
import { Badge, type BadgeProps } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

interface EnvironmentStatusIndicatorProps {
  environmentName: string
  initialState?: string
  className?: string
  showLoader?: boolean
}

function getStatusVariant(state: string | undefined): BadgeProps['variant'] {
  switch (state) {
    case 'active':
      return 'success'
    case 'inactive':
      return 'secondary'
    case 'activating':
      return 'info'
    case 'deactivating':
      return 'warning'
    case 'deleting':
    case 'error':
      return 'destructive'
    default:
      return 'secondary'
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
    <Badge
      variant={getStatusVariant(displayState)}
      className={cn('gap-1', className)}
    >
      {showLoader && isTransitionalState(displayState) && (
        <Loader2 className="h-3 w-3 animate-spin" />
      )}
      {displayState || 'unknown'}
    </Badge>
  )
}
