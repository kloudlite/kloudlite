'use client'

import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface WorkspaceStatusIndicatorProps {
  phase?: string
  className?: string
  showLoader?: boolean
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
  phase,
  className,
  showLoader = true,
}: WorkspaceStatusIndicatorProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium',
        getPhaseStyles(phase),
        className
      )}
    >
      {showLoader && isTransitionalPhase(phase) && (
        <Loader2 className="h-3 w-3 animate-spin" />
      )}
      {phase || 'Unknown'}
    </span>
  )
}
