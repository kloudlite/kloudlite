import { cn } from '@/lib/utils'

interface DocContainerProps {
  children: React.ReactNode
  className?: string
}

export function DocContainer({ children, className }: DocContainerProps) {
  return (
    <div
      className={cn(
        'prose prose-slate dark:prose-invert mx-auto max-w-3xl px-4 pt-8 pb-16 sm:px-6 lg:px-8 xl:pr-16',
        className
      )}
    >
      {children}
    </div>
  )
}
