import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

type Variant =
  | 'default'
  | 'secret'
  | 'config'
  | 'success'
  | 'danger'
  | 'warning'
  | 'info'
  | 'muted'

const variants: Record<Variant, string> = {
  default: 'bg-accent text-muted-foreground',
  secret: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
  config: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  success: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
  danger: 'bg-red-500/10 text-red-600 dark:text-red-400',
  warning: 'bg-amber-500/15 text-amber-500',
  info: 'bg-blue-500/15 text-blue-500',
  muted: 'bg-sidebar-foreground/10 text-sidebar-foreground/50',
}

interface BadgeProps {
  variant?: Variant
  children: ReactNode
  className?: string
}

export function Badge({ variant = 'default', children, className }: BadgeProps) {
  return (
    <span className={cn(
      'shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium',
      variants[variant],
      className
    )}>
      {children}
    </span>
  )
}
