import * as React from "react"
import { Button, type ButtonProps } from "./button"
import { cn } from "@/lib/utils"

interface IconButtonProps extends Omit<ButtonProps, 'size' | 'children'> {
  children: React.ReactNode
  size?: "default" | "sm" | "lg"
  label: string // Required for accessibility
  showTooltip?: boolean
}

const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ className, children, size = "default", label, variant = "outline", ...props }, ref) => {
    const sizeMap = {
      sm: "icon-sm",
      default: "icon",
      lg: "icon-lg"
    } as const

    return (
      <Button
        ref={ref}
        size={sizeMap[size]}
        variant={variant}
        className={cn(
          "relative overflow-hidden",
          className
        )}
        aria-label={label}
        {...props}
      >
        {children}
      </Button>
    )
  }
)

IconButton.displayName = "IconButton"

export { IconButton }