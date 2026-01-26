import { ReactNode } from 'react'
import { LucideIcon } from 'lucide-react'

interface FeatureCardProps {
  icon: LucideIcon
  title: string | ReactNode
  description: string | ReactNode
}

export function FeatureCard({ icon: Icon, title, description }: FeatureCardProps) {
  return (
    <div className="group relative overflow-hidden bg-foreground/[0.015] border border-foreground/10 hover:bg-foreground/[0.03] hover:border-foreground/20 transition-[background-color,border-color] duration-300 p-6 rounded-sm">
      {/* Animated vertical accent bar */}
      <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />

      <div className="flex items-start gap-4 relative z-10">
        <div className="flex-shrink-0 w-10 h-10 rounded-sm bg-primary/10 border border-primary/20 flex items-center justify-center group-hover:bg-primary/15 transition-[background-color,transform] duration-300 group-hover:scale-110">
          <Icon className="h-5 w-5 text-primary" />
        </div>
        <div className="flex-1">
          <h3 className="text-foreground group-hover:text-primary text-base font-semibold mb-2 tracking-tight transition-colors duration-300">{title}</h3>
          <div className="text-muted-foreground group-hover:text-foreground text-[14px] leading-relaxed transition-colors duration-300">
            {description}
          </div>
        </div>
      </div>
    </div>
  )
}
