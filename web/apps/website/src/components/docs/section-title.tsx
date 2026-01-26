import { ReactNode } from 'react'

interface SectionTitleProps {
  children: ReactNode
  id?: string
}

export function SectionTitle({ children, id }: SectionTitleProps) {
  return (
    <div className="relative mt-16 mb-6 border-t border-foreground/10 pt-8 group">
      <h2 id={id} className="text-foreground font-bold tracking-tight text-2xl lg:text-3xl leading-tight relative inline-block">
        {children}
        {/* Animated underline on hover */}
        <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
      </h2>
    </div>
  )
}
