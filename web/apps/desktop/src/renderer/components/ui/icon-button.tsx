import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface IconButtonProps {
  children: ReactNode
  onClick?: (e: React.MouseEvent) => void
  variant?: 'muted' | 'danger'
  size?: 'sm' | 'md'
  className?: string
  disabled?: boolean
}

export function IconButton({ children, onClick, variant = 'muted', size = 'sm', className, disabled }: IconButtonProps) {
  return (
    <button
      className={cn(
        'flex shrink-0 items-center justify-center rounded transition-colors',
        size === 'sm' ? 'h-6 w-6' : 'h-7 w-7',
        variant === 'muted' && 'text-muted-foreground/40 hover:bg-accent hover:text-muted-foreground',
        variant === 'danger' && 'text-muted-foreground/40 hover:bg-red-500/10 hover:text-red-500',
        disabled && 'pointer-events-none opacity-40',
        className
      )}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  )
}
