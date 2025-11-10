import { cn } from '@/lib/utils'
import { LucideIcon } from 'lucide-react'

interface FeatureCardProps {
  icon: LucideIcon
  title: string
  description: string
  variant?: 'primary' | 'destructive'
  className?: string
}

export function FeatureCard({
  icon: Icon,
  title,
  description,
  variant = 'primary',
  className,
}: FeatureCardProps) {
  return (
    <div className={cn('bg-card rounded-lg border p-4 sm:p-6', className)}>
      <div
        className={cn(
          'mb-3 flex h-12 w-12 items-center justify-center rounded-lg',
          variant === 'primary' && 'bg-primary',
          variant === 'destructive' && 'bg-destructive'
        )}
      >
        <Icon
          className={cn(
            'h-6 w-6',
            variant === 'primary' && 'text-primary-foreground',
            variant === 'destructive' && 'text-destructive-foreground'
          )}
        />
      </div>
      <h3 className="text-card-foreground mb-2 text-xl font-semibold leading-snug">{title}</h3>
      <p className="text-muted-foreground leading-relaxed">{description}</p>
    </div>
  )
}
