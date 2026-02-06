'use client'

import { Loader2 } from 'lucide-react'
import { Badge, type BadgeProps } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import { useResourceWatch } from '@/lib/hooks/use-resource-watch'

interface WorkspaceStatusIndicatorProps {
  phase?: string
  className?: string
  showLoader?: boolean
}

function getPhaseVariant(phase: string | undefined): BadgeProps['variant'] {
  switch (phase) {
    case 'Running':
      return 'success'
    case 'Stopped':
      return 'secondary'
    case 'Pending':
    case 'Creating':
      return 'info'
    case 'Stopping':
      return 'warning'
    case 'Terminating':
    case 'Failed':
      return 'destructive'
    default:
      return 'secondary'
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
  useResourceWatch('workspaces')

  return (
    <Badge
      variant={getPhaseVariant(phase)}
      className={cn('gap-1', className)}
    >
      {showLoader && isTransitionalPhase(phase) && (
        <Loader2 className="h-3 w-3 animate-spin" />
      )}
      {phase || 'Unknown'}
    </Badge>
  )
}
