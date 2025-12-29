import { cn } from '@/lib/utils'

interface DocContainerProps {
  children: React.ReactNode
  className?: string
}

export function DocContainer({ children, className }: DocContainerProps) {
  return (
    <div
      className={cn(
        'mx-auto max-w-3xl px-4 pt-6 pb-24 sm:pt-8 sm:px-6 lg:px-8 lg:pb-16 xl:pr-16',
        className
      )}
    >
      {children}
    </div>
  )
}
