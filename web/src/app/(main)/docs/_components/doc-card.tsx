import { cn } from '@/lib/utils'

interface DocCardProps {
  children: React.ReactNode
  className?: string
}

export function DocCard({ children, className }: DocCardProps) {
  return <div className={cn('bg-card rounded-lg border p-4 sm:p-6', className)}>{children}</div>
}

interface DocCardTitleProps {
  children: React.ReactNode
  className?: string
}

export function DocCardTitle({ children, className }: DocCardTitleProps) {
  return (
    <h3 className={cn('text-card-foreground mb-3 text-xl font-semibold leading-snug', className)}>
      {children}
    </h3>
  )
}

interface DocCardDescriptionProps {
  children: React.ReactNode
  className?: string
}

export function DocCardDescription({ children, className }: DocCardDescriptionProps) {
  return <p className={cn('text-muted-foreground leading-relaxed', className)}>{children}</p>
}
