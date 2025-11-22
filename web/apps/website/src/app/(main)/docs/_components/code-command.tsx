import { cn } from '@/lib/utils'

interface CodeCommandProps {
  children: React.ReactNode
  className?: string
}

export function CodeCommand({ children, className }: CodeCommandProps) {
  return <div className={cn('bg-muted rounded p-3 font-mono text-sm', className)}>{children}</div>
}

interface CodeInlineProps {
  children: React.ReactNode
  className?: string
}

export function CodeInline({ children, className }: CodeInlineProps) {
  return (
    <code className={cn('bg-muted rounded px-1.5 py-0.5 font-mono text-sm', className)}>
      {children}
    </code>
  )
}
