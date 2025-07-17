import { cn } from '@/lib/utils'
import { cva, type VariantProps } from 'class-variance-authority'

const infoBoxVariants = cva(
  "rounded-lg p-6 my-8 border",
  {
    variants: {
      variant: {
        default: "bg-muted/50 border border-border",
        primary: "bg-primary/5 border-primary/20",
        success: "bg-success/5 border-success/20", 
        warning: "bg-warning/5 border-warning/20",
        destructive: "bg-destructive/5 border-destructive/20",
        info: "bg-info/5 border-info/20"
      }
    },
    defaultVariants: {
      variant: "default"
    }
  }
)

interface InfoBoxProps extends VariantProps<typeof infoBoxVariants> {
  title?: string
  children: React.ReactNode
  className?: string
}

export function InfoBox({ 
  title,
  variant,
  children, 
  className 
}: InfoBoxProps) {
  return (
    <div className={cn(infoBoxVariants({ variant }), className)}>
      {title && (
        <h3 className={cn(
          "text-lg font-semibold mb-4",
          variant === "primary" && "text-primary",
          variant === "success" && "text-success",
          variant === "warning" && "text-warning",
          variant === "destructive" && "text-destructive",
          variant === "info" && "text-info"
        )}>
          {title}
        </h3>
      )}
      {children}
    </div>
  )
}