import * as React from "react"

import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "bg-primary/10 text-primary hover:bg-primary/20",
        secondary:
          "bg-muted text-muted-foreground hover:bg-muted/80",
        destructive:
          "bg-destructive/10 text-destructive hover:bg-destructive/20",
        outline:
          "border border-input text-foreground hover:bg-accent hover:text-accent-foreground",
        success:
          "bg-green-500/10 text-green-600 dark:text-green-500 hover:bg-green-500/20",
        warning:
          "bg-yellow-500/10 text-yellow-600 dark:text-yellow-500 hover:bg-yellow-500/20",
        error:
          "bg-red-500/10 text-red-600 dark:text-red-500 hover:bg-red-500/20",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span"

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
