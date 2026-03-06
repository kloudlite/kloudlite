import { cn } from '@kloudlite/lib'

interface CrossMarkerProps {
  className?: string
}

export function CrossMarker({ className }: CrossMarkerProps) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 h-5 w-px -translate-x-1/2 bg-foreground/20" />
      <div className="absolute left-0 top-1/2 h-px w-5 -translate-y-1/2 bg-foreground/20" />
    </div>
  )
}
