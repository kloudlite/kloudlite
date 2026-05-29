import { cn } from '@/lib/utils'

type Status = 'active' | 'running' | 'error' | 'failed' | 'stopped' | 'inactive' | 'starting' | 'idle'

const colors: Record<Status, string> = {
  active: 'bg-emerald-400',
  running: 'bg-emerald-400',
  error: 'bg-red-400',
  failed: 'bg-red-400',
  stopped: 'bg-sidebar-foreground/25',
  inactive: 'bg-sidebar-foreground/25',
  starting: 'bg-amber-400',
  idle: 'bg-amber-400',
}

interface StatusDotProps {
  status: Status
  size?: 'sm' | 'md'
  className?: string
}

export function StatusDot({ status, size = 'sm', className }: StatusDotProps) {
  return (
    <div
      className={cn(
        'shrink-0 rounded-full',
        size === 'sm' ? 'h-2 w-2' : 'h-2.5 w-2.5',
        colors[status],
        className
      )}
    />
  )
}
