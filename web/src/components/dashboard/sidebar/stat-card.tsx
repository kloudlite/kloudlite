import { cn } from '@/lib/utils'
import { cva, type VariantProps } from 'class-variance-authority'
import { LucideIcon } from 'lucide-react'

const statCardVariants = cva(
  "bg-background/60 backdrop-blur-sm border border-border/40 rounded-lg py-3 px-2 text-center hover:bg-background/80 transition-all duration-200 cursor-default group",
  {
    variants: {
      type: {
        cpu: "[&_.icon-wrapper]:bg-primary/10 [&_.icon-wrapper]:group-hover:bg-primary/20 [&_.icon]:text-primary",
        memory: "[&_.icon-wrapper]:bg-accent/10 [&_.icon-wrapper]:group-hover:bg-accent/20 [&_.icon]:text-accent",
        uptime: "[&_.icon-wrapper]:bg-warning/10 [&_.icon-wrapper]:group-hover:bg-warning/20 [&_.icon]:text-warning"
      }
    }
  }
)

interface StatCardProps extends VariantProps<typeof statCardVariants> {
  icon: LucideIcon
  label: string
  value: string
  className?: string
}

export function StatCard({ type, icon: Icon, label, value, className }: StatCardProps) {
  return (
    <div className={cn(statCardVariants({ type }), className)}>
      <div className="flex flex-col items-center gap-1.5">
        <div className="icon-wrapper p-1.5 rounded-md transition-colors">
          <Icon className="icon size-3.5" />
        </div>
        <span className="text-xs font-medium text-muted-foreground">{label}</span>
        <p className="text-sm font-bold text-foreground">{value}</p>
      </div>
    </div>
  )
}