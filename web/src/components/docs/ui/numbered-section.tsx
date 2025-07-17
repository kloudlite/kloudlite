import { cn } from '@/lib/utils'

interface NumberedSectionProps {
  number: number
  title: string
  children: React.ReactNode
  className?: string
}

export function NumberedSection({ 
  number, 
  title, 
  children, 
  className 
}: NumberedSectionProps) {
  return (
    <div className={cn("numbered-section", className)}>
      <div className="number-circle">
        {number}
      </div>
      <div className="content">
        <h3>{title}</h3>
        {children}
      </div>
    </div>
  )
}