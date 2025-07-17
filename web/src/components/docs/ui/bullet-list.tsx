import { cn } from '@/lib/utils'
import { cva, type VariantProps } from 'class-variance-authority'

const bulletListVariants = cva(
  "space-y-2",
  {
    variants: {
      variant: {
        default: "",
        compact: "text-sm"
      },
      spacing: {
        default: "ml-6",
        none: "",
        tight: "ml-4",
        loose: "ml-8"
      }
    },
    defaultVariants: {
      variant: "default",
      spacing: "default"
    }
  }
)

const bulletVariants = cva(
  "",
  {
    variants: {
      color: {
        default: "text-muted-foreground",
        primary: "text-primary",
        success: "text-success",
        warning: "text-warning",
        destructive: "text-destructive"
      }
    },
    defaultVariants: {
      color: "default"
    }
  }
)

interface BulletListProps extends VariantProps<typeof bulletListVariants> {
  items: string[]
  bulletColor?: VariantProps<typeof bulletVariants>['color']
  className?: string
}

export function BulletList({ 
  items,
  variant,
  spacing,
  bulletColor = "default",
  className 
}: BulletListProps) {
  return (
    <ul className={cn(bulletListVariants({ variant, spacing }), className)}>
      {items.map((item, index) => (
        <li key={index} className="flex items-start gap-2">
          <span className={cn(bulletVariants({ color: bulletColor }))}>•</span>
          <span>{item}</span>
        </li>
      ))}
    </ul>
  )
}