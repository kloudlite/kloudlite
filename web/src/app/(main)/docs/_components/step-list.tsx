import { cn } from '@/lib/utils'

interface StepListProps {
  children: React.ReactNode
  className?: string
}

export function StepList({ children, className }: StepListProps) {
  return <ol className={cn('space-y-4', className)}>{children}</ol>
}

interface StepItemProps {
  number: number
  title: string
  children: React.ReactNode
  className?: string
}

export function StepItem({ number, title, children, className }: StepItemProps) {
  return (
    <li className={cn('flex items-start gap-3', className)}>
      <span className="bg-primary text-primary-foreground mt-1 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-sm font-semibold">
        {number}
      </span>
      <div className="flex-1">
        <p className="text-card-foreground font-medium leading-snug">{title}</p>
        {children}
      </div>
    </li>
  )
}
