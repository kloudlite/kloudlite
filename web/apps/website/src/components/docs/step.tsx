import { ReactNode } from 'react'

interface StepProps {
  number: number
  title: string
  description: string
  children?: ReactNode
}

export function Step({ number, title, description, children }: StepProps) {
  return (
    <div className="group relative overflow-hidden flex items-start gap-6 p-6 rounded-sm hover:bg-foreground/[0.02] transition-[background-color] duration-300">
      {/* Animated vertical accent bar */}
      <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />

      <div className="flex-shrink-0 w-12 h-12 rounded-sm bg-primary/10 border border-primary/20 flex items-center justify-center relative z-10">
        <span className="text-primary font-bold text-xl">{number}</span>
      </div>
      <div className="flex-1 pt-1 relative z-10">
        <h3 className="text-foreground group-hover:text-primary text-lg font-semibold mb-2 tracking-tight transition-colors duration-300">{title}</h3>
        <p className="text-muted-foreground group-hover:text-foreground text-[15px] leading-relaxed mb-4 transition-colors duration-300">{description}</p>
        {children}
      </div>
    </div>
  )
}
