import { cn } from '@/lib/utils'

interface FeatureGridProps {
  children: React.ReactNode
  columns?: 2 | 3
  className?: string
}

export function FeatureGrid({ children, columns = 2, className }: FeatureGridProps) {
  return (
    <div
      className={cn(
        'grid gap-4 sm:gap-6',
        columns === 2 && 'md:grid-cols-2',
        columns === 3 && 'md:grid-cols-3',
        className
      )}
    >
      {children}
    </div>
  )
}
