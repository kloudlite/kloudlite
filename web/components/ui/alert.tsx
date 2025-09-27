import * as React from "react"

import { cva, type VariantProps } from "class-variance-authority"
import { AlertCircle, CheckCircle2, Info, XCircle } from "lucide-react"

import { cn } from "@/lib/utils"

const alertVariants = cva(
  "relative w-full rounded-lg border p-4 flex gap-3 items-start transition-all duration-200 shadow-sm",
  {
    variants: {
      variant: {
        default: "bg-card border-border text-card-foreground",
        info: "bg-primary/10 border-primary/20 text-primary dark:bg-primary/15 dark:border-primary/30",
        success: "bg-emerald-50 border-emerald-200 text-emerald-900 dark:bg-emerald-950/30 dark:border-emerald-800 dark:text-emerald-100",
        warning: "bg-amber-50 border-amber-200 text-amber-900 dark:bg-amber-950/30 dark:border-amber-800 dark:text-amber-100",
        destructive: "bg-destructive/10 border-destructive/20 text-destructive dark:bg-destructive/15 dark:border-destructive/30",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

const iconMap = {
  default: Info,
  info: Info,
  success: CheckCircle2,
  warning: AlertCircle,
  destructive: XCircle,
}

interface AlertProps extends React.ComponentProps<"div">, VariantProps<typeof alertVariants> {
  icon?: React.ReactNode
}

function Alert({
  className,
  variant = "default",
  icon,
  children,
  ...props
}: AlertProps) {
  const Icon = iconMap[variant || "default"]
  const alertIcon = icon || <Icon className="h-5 w-5 flex-shrink-0" />
  
  return (
    <div
      data-slot="alert"
      role="alert"
      className={cn(alertVariants({ variant }), className)}
      {...props}
    >
      <div className="flex-shrink-0">
        {alertIcon}
      </div>
      <div className="flex-1 space-y-1">
        {children}
      </div>
    </div>
  )
}

function AlertTitle({ className, ...props }: React.ComponentProps<"h5">) {
  return (
    <h5
      data-slot="alert-title"
      className={cn(
        "font-semibold text-base leading-none tracking-tight",
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
        "text-sm leading-relaxed opacity-90 first-letter:uppercase",
        className
      )}
      {...props}
    />
  )
}

export { Alert, AlertTitle, AlertDescription }
