import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-none text-sm font-medium transition-all duration-200 ease-in-out disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 [&_svg]:transition-transform [&_svg]:duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background active:scale-[0.99] active:transition-transform active:duration-75",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground hover:bg-primary-hover active:bg-primary-active focus-visible:ring-primary",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive-hover active:bg-destructive-active focus-visible:ring-destructive",
        outline:
          "border border-input bg-background hover:border-primary hover:text-primary active:border-primary-active focus-visible:ring-ring",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary-hover active:bg-secondary-active focus-visible:ring-ring",
        ghost:
          "hover:bg-muted hover:text-primary active:bg-muted/80 focus-visible:ring-ring",
        link: "text-primary underline-offset-4 hover:underline hover:text-primary-hover active:text-primary-active focus-visible:ring-ring",
        success:
          "bg-success text-success-foreground hover:bg-success-hover active:bg-success-active focus-visible:ring-success",
        warning:
          "bg-warning text-warning-foreground hover:bg-warning-hover active:bg-warning-active focus-visible:ring-warning",
        info:
          "bg-info text-info-foreground hover:bg-info-hover active:bg-info-active focus-visible:ring-info",
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        sm: "h-8 gap-1.5 px-3 text-xs has-[>svg]:px-2.5",
        lg: "h-12 px-8 text-base gap-3 has-[>svg]:px-5",
        xl: "h-12 px-10 text-lg gap-3 has-[>svg]:px-6",
        auth: "h-11 px-5 text-base gap-2.5 has-[>svg]:px-4",
        icon: "size-9 hover:[&_svg]:rotate-12",
        "icon-sm": "size-8 hover:[&_svg]:rotate-12",
        "icon-lg": "size-11 hover:[&_svg]:rotate-12",
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
