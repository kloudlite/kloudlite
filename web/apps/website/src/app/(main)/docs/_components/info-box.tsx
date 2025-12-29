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
        'border-l-4 p-3 sm:p-4',
        variant === 'note' && 'bg-muted/50 border-muted-foreground/40',
        variant === 'warning' && 'bg-destructive/10 border-destructive',
        className
      )}
    >
      <div className="text-card-foreground text-sm leading-relaxed">{children}</div>
    </div>
  )
}
