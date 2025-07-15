import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-none border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background aria-invalid:ring-2 aria-invalid:ring-destructive aria-invalid:border-destructive transition-all duration-200 overflow-hidden",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground focus-visible:ring-ring [[href]&]:hover:bg-primary-hover [[href]&]:active:bg-primary-active [button&]:hover:bg-primary-hover [button&]:active:bg-primary-active",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground focus-visible:ring-ring [[href]&]:hover:bg-muted [[href]&]:active:bg-secondary-active [button&]:hover:bg-muted [button&]:active:bg-secondary-active",
        destructive:
          "border-transparent bg-destructive text-destructive-foreground focus-visible:ring-destructive [[href]&]:hover:bg-destructive-hover [[href]&]:active:bg-destructive-active [button&]:hover:bg-destructive-hover [button&]:active:bg-destructive-active",
        outline:
          "border-border bg-background text-foreground focus-visible:ring-ring [[href]&]:hover:bg-muted [[href]&]:active:bg-secondary [button&]:hover:bg-muted [button&]:active:bg-secondary",
        success:
          "border-transparent bg-success text-success-foreground focus-visible:ring-success [[href]&]:hover:bg-success-hover [[href]&]:active:bg-success-active [button&]:hover:bg-success-hover [button&]:active:bg-success-active",
        warning:
          "border-transparent bg-warning text-warning-foreground focus-visible:ring-warning [[href]&]:hover:bg-warning-hover [[href]&]:active:bg-warning-active [button&]:hover:bg-warning-hover [button&]:active:bg-warning-active",
        info:
          "border-transparent bg-info text-info-foreground focus-visible:ring-info [[href]&]:hover:bg-info-hover [[href]&]:active:bg-info-active [button&]:hover:bg-info-hover [button&]:active:bg-info-active",
        "success-subtle":
          "border-success-border bg-success-background text-success-text focus-visible:ring-success [[href]&]:hover:bg-success-subtle-hover [[href]&]:active:bg-success-subtle-active [button&]:hover:bg-success-subtle-hover [button&]:active:bg-success-subtle-active",
        "warning-subtle":
          "border-warning-border bg-warning-background text-warning-text focus-visible:ring-warning [[href]&]:hover:bg-warning-subtle-hover [[href]&]:active:bg-warning-subtle-active [button&]:hover:bg-warning-subtle-hover [button&]:active:bg-warning-subtle-active",
        "info-subtle":
          "border-info-border bg-info-background text-info-text focus-visible:ring-info [[href]&]:hover:bg-info-subtle-hover [[href]&]:active:bg-info-subtle-active [button&]:hover:bg-info-subtle-hover [button&]:active:bg-info-subtle-active",
        "destructive-subtle":
          "border-destructive-border bg-destructive-background text-destructive-text focus-visible:ring-destructive [[href]&]:hover:bg-destructive-subtle-hover [[href]&]:active:bg-destructive-subtle-active [button&]:hover:bg-destructive-subtle-hover [button&]:active:bg-destructive-subtle-active",
      },
      interactive: {
        true: "cursor-pointer select-none",
        false: "",
      },
    },
    defaultVariants: {
      variant: "default",
      interactive: false,
    },
  }
)

function Badge({
  className,
  variant,
  interactive,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span"

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant, interactive }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
