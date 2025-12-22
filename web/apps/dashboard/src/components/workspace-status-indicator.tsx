'use client'

import { useWorkspaceStatusStream } from '@/lib/hooks/use-workspace-status-stream'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

interface WorkspaceStatusIndicatorProps {
  namespace: string
  workspaceName: string
  initialPhase?: string
  className?: string
  showLoader?: boolean
  onReady?: () => void
}

function getPhaseStyles(phase: string | undefined) {
  switch (phase) {
    case 'Running':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
    case 'Stopped':
      return 'bg-secondary text-secondary-foreground'
    case 'Pending':
    case 'Creating':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    case 'Stopping':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
    case 'Terminating':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    case 'Failed':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-secondary text-secondary-foreground'
  }
}

function isTransitionalPhase(phase: string | undefined) {
  return phase === 'Pending' || phase === 'Creating' || phase === 'Stopping' || phase === 'Terminating'
}

export function WorkspaceStatusIndicator({
  namespace,
  workspaceName,
  initialPhase,
  className,
  showLoader = true,
  onReady,
}: WorkspaceStatusIndicatorProps) {
  const router = useRouter()
  const { phase, isConnected } = useWorkspaceStatusStream(namespace, workspaceName, {
    enabled: true,
    onPhaseChange: (newPhase) => {
      // Refresh the page when transitioning to a stable state
      if (!isTransitionalPhase(newPhase)) {
        router.refresh()
      }
    },
    onReady,
  })

  // Use SSE phase if connected, otherwise fall back to initial phase
  const displayPhase = isConnected && phase ? phase : initialPhase

  // Refresh page data when phase changes to ensure consistency
  useEffect(() => {
    if (isConnected && phase && phase !== initialPhase) {
      router.refresh()
    }
  }, [isConnected, phase, initialPhase, router])

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium',
        getPhaseStyles(displayPhase),
        className
      )}
    >
      {showLoader && isTransitionalPhase(displayPhase) && (
        <Loader2 className="h-3 w-3 animate-spin" />
      )}
      {displayPhase || 'Unknown'}
    </span>
  )
}
