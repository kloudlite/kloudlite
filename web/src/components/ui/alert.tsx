import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const alertVariants = cva(
  "relative w-full rounded-none border px-4 py-4 text-sm grid has-[>svg]:grid-cols-[calc(var(--spacing)*6)_1fr] grid-cols-[0_1fr] has-[>svg]:gap-x-3 gap-y-1 items-start [&>svg]:size-5 [&>svg]:text-current [&>svg]:transition-colors [&>svg]:duration-200",
  {
    variants: {
      variant: {
        default: "bg-card text-card-foreground border-border",
        destructive:
          "bg-destructive-background text-destructive-text border-destructive-border [&>svg]:text-destructive-text [&_[data-slot=alert-description]]:text-destructive-text-muted",
        success:
          "bg-success-background text-success-text border-success-border [&>svg]:text-success-text [&_[data-slot=alert-description]]:text-success-text-muted",
        warning:
          "bg-warning-background text-warning-text border-warning-border [&>svg]:text-warning-text [&_[data-slot=alert-description]]:text-warning-text-muted",
        info:
          "bg-info-background text-info-text border-info-border [&>svg]:text-info-text [&_[data-slot=alert-description]]:text-info-text-muted",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Alert({
  className,
  variant,
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof alertVariants>) {
  return (
    <div
      data-slot="alert"
      role="alert"
      className={cn(alertVariants({ variant }), className)}
      {...props}
    />
  )
}

function AlertTitle({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-title"
      className={cn(
        "col-start-2 text-base font-semibold tracking-tight leading-none",
        className
      )}
      {...props}
    />
  )
}

function AlertDescription({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-description"
      className={cn(
        "text-muted-foreground col-start-2 grid justify-items-start gap-1 text-sm [&_p]:leading-relaxed",
        className
      )}
      {...props}
    />
  )
}

export { Alert, AlertTitle, AlertDescription }
