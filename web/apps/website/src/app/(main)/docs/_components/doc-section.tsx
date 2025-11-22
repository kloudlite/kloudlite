import { cn } from '@/lib/utils'

interface DocSectionProps {
  title: string
  children: React.ReactNode
  className?: string
}

export function DocSection({ title, children, className }: DocSectionProps) {
  return (
    <section className={cn('mb-12 sm:mb-16', className)}>
      <h2 className="text-foreground mb-4 sm:mb-6 text-2xl sm:text-3xl font-bold">{title}</h2>
      {children}
    </section>
  )
}
