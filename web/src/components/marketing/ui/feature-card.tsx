import { cn } from '@/lib/utils'
import { LucideIcon } from 'lucide-react'
import { cva, type VariantProps } from 'class-variance-authority'

const featureCardVariants = cva(
  "group relative p-4 sm:p-6 border transition-all duration-300 rounded-lg",
  {
    variants: {
      variant: {
        default: "border-border hover:border-primary/50 hover:shadow-lg hover:-translate-y-1",
        primary: "border-primary bg-primary/5 hover:bg-primary/10 hover:shadow-lg hover:-translate-y-1",
        ghost: "border-transparent hover:border-border hover:bg-muted/50"
      }
    },
    defaultVariants: {
      variant: "default"
    }
  }
)

interface FeatureCardProps extends VariantProps<typeof featureCardVariants> {
  icon: LucideIcon
  title: string
  description: string
  iconClassName?: string
  className?: string
}

export function FeatureCard({ 
  icon: Icon,
  title,
  description,
  variant,
  iconClassName,
  className 
}: FeatureCardProps) {
  return (
    <div className={cn(featureCardVariants({ variant }), className)}>
      <div className="absolute inset-0 bg-gradient-to-br from-primary/0 to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
      <Icon className={cn(
        "h-6 w-6 sm:h-8 sm:w-8 text-primary mb-3 sm:mb-4 transition-all duration-300",
        iconClassName
      )} />
      <h3 className="font-semibold mb-2 text-sm sm:text-base group-hover:text-primary transition-colors">
        {title}
      </h3>
      <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed">
        {description}
      </p>
    </div>
  )
}