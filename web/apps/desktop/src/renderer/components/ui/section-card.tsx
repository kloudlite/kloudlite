import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface SectionCardProps {
  title: string
  description?: string
  variant?: 'default' | 'danger'
  children: ReactNode
  className?: string
}

export function SectionCard({ title, description, variant = 'default', children, className }: SectionCardProps) {
  return (
    <div className={cn(
      'rounded-xl border bg-card p-5',
      variant === 'danger' ? 'border-red-500/20' : 'border-border/50',
      className
    )}>
      <h3 className={cn(
        'text-[13px] font-semibold',
        variant === 'danger' ? 'text-red-500' : 'text-foreground'
      )}>
        {title}
      </h3>
      {description && (
        <p className="mt-1 text-[12px] text-muted-foreground">{description}</p>
      )}
      <div className="mt-3">{children}</div>
    </div>
  )
}
