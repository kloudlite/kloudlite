import { cn } from '@/lib/utils'

interface DocHeaderProps {
  title: string
  description: string
  className?: string
}

export function DocHeader({ title, description, className }: DocHeaderProps) {
  return (
    <div className={cn('mb-12 sm:mb-16', className)}>
      <h1 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl lg:text-5xl break-words leading-tight sm:leading-tight">
        {title}
      </h1>
      <p className="text-muted-foreground mt-4 text-base sm:text-lg lg:text-xl leading-relaxed">
        {description}
      </p>
    </div>
  )
}
