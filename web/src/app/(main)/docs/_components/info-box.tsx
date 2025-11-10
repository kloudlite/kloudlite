import { cn } from '@/lib/utils'

interface InfoBoxProps {
  children: React.ReactNode
  variant?: 'note' | 'warning'
  className?: string
}

export function InfoBox({ children, variant = 'note', className }: InfoBoxProps) {
  return (
    <div
      className={cn(
        'rounded-lg border p-3 sm:p-4',
        variant === 'note' && 'bg-muted/50 border-muted-foreground/20',
        variant === 'warning' && 'bg-destructive/10 border-destructive/50',
        className
      )}
    >
      <div className="text-card-foreground text-sm leading-relaxed">{children}</div>
    </div>
  )
}
