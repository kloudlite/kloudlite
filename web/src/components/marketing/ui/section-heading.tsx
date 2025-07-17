import { cn } from '@/lib/utils'

interface SectionHeadingProps {
  title: string | React.ReactNode
  subtitle?: string | React.ReactNode
  align?: 'left' | 'center' | 'right'
  className?: string
}

export function SectionHeading({ 
  title, 
  subtitle, 
  align = 'center',
  className 
}: SectionHeadingProps) {
  return (
    <div className={cn(
      "space-y-2",
      align === 'center' && "text-center",
      align === 'left' && "text-left",
      align === 'right' && "text-right",
      className
    )}>
      <h2 className="text-3xl font-bold">
        {title}
      </h2>
      {subtitle && (
        <p className="text-2xl text-primary font-semibold">
          {subtitle}
        </p>
      )}
    </div>
  )
}