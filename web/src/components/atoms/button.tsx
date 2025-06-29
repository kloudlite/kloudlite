import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  cn(
    "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md",
    "text-sm font-medium transition-all duration-200",
    "disabled:pointer-events-none disabled:opacity-50",
    "[&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
    // Professional click feedback
    "active:scale-[0.98] active:transition-transform active:duration-75"
  ),
  {
    variants: {
      variant: {
        default: cn(
          "bg-primary text-primary-foreground shadow-button",
          "hover:bg-primary/90 hover:shadow-button-hover",
          "active:bg-primary active:shadow-sm"
        ),
        destructive: cn(
          "bg-destructive text-destructive-foreground shadow-button",
          "hover:bg-destructive/90 hover:shadow-button-hover",
          "active:bg-destructive active:shadow-sm"
        ),
        outline: cn(
          "border border-border bg-transparent",
          "hover:bg-muted/50 hover:border-muted-foreground/25",
          "active:bg-muted active:border-muted-foreground/30"
        ),
        secondary: cn(
          "bg-secondary text-secondary-foreground border border-transparent",
          "hover:bg-secondary/80 hover:border-border/50",
          "active:bg-secondary active:border-border"
        ),
        ghost: cn(
          "hover:bg-accent hover:text-accent-foreground",
          "active:bg-accent/80"
        ),
        link: cn(
          "text-link underline-offset-4 hover:underline",
          "active:text-link/80"
        ),
      },
      size: {
        sm: "h-8 px-3 text-xs",
        default: "h-10 px-4 text-sm",
        lg: "h-12 px-6 text-base",
        icon: "h-10 w-10 p-0",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot : "button"

  return (
    <Comp
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
