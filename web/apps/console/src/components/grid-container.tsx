import { cn } from '@kloudlite/lib'

export function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative', className)}>
      {children}
    </div>
  )
}
