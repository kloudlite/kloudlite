import { cn } from '@/lib/utils'
import { cva, type VariantProps } from 'class-variance-authority'

const teamAvatarVariants = cva(
  "rounded-lg flex items-center justify-center font-semibold bg-gradient-to-br from-primary to-primary/90 text-primary-foreground",
  {
    variants: {
      size: {
        sm: "size-8 text-xs",
        md: "size-10 text-sm",
        lg: "size-12 text-base"
      }
    },
    defaultVariants: {
      size: "md"
    }
  }
)

const statusIndicatorVariants = cva(
  "absolute rounded-full",
  {
    variants: {
      size: {
        sm: "-bottom-0.5 -right-0.5 size-2",
        md: "-bottom-1 -right-1 size-2.5",
        lg: "-bottom-1 -right-1 size-3"
      },
      status: {
        active: "bg-success",
        inactive: "bg-destructive"
      }
    },
    defaultVariants: {
      size: "md",
      status: "active"
    }
  }
)

interface TeamAvatarProps extends VariantProps<typeof teamAvatarVariants> {
  name: string
  showStatus?: boolean
  status?: 'active' | 'inactive'
  className?: string
}

export function TeamAvatar({ 
  name, 
  size, 
  showStatus = false, 
  status = 'active',
  className 
}: TeamAvatarProps) {
  const initials = name
    .split(' ')
    .map(word => word.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <div className="relative">
      <div className={cn(teamAvatarVariants({ size }), className)}>
        {initials}
      </div>
      {showStatus && (
        <div className={statusIndicatorVariants({ size, status })} />
      )}
    </div>
  )
}