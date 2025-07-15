"use client"

import * as React from "react"
import * as TogglePrimitive from "@radix-ui/react-toggle"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const toggleVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all duration-200 ease-in-out disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 [&_svg]:transition-transform [&_svg]:duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background active:scale-[0.98] active:transition-[transform,box-shadow] active:duration-100",
  {
    variants: {
      variant: {
        default: [
          "bg-muted/50",
          "hover:bg-muted hover:text-muted-foreground",
          "data-[state=on]:bg-primary data-[state=on]:text-primary-foreground",
          "data-[state=on]:shadow-sm",
          "data-[state=on]:hover:bg-primary-hover",
          "focus-visible:ring-ring",
          "[&_svg]:transition-transform [&_svg]:duration-200",
          "data-[state=on]:[&_svg]:scale-110",
        ],
        outline: [
          "border border-input bg-background",
          "hover:border-primary hover:text-primary",
          "data-[state=on]:bg-primary data-[state=on]:text-primary-foreground data-[state=on]:border-primary",
          "data-[state=on]:shadow-sm",
          "data-[state=on]:hover:bg-primary-hover",
          "focus-visible:ring-ring",
          "[&_svg]:transition-transform [&_svg]:duration-200",
          "data-[state=on]:[&_svg]:scale-110",
        ],
        secondary: [
          "bg-secondary text-secondary-foreground",
          "hover:bg-secondary/80",
          "data-[state=on]:bg-primary data-[state=on]:text-primary-foreground",
          "data-[state=on]:shadow-sm",
          "data-[state=on]:hover:bg-primary-hover",
          "focus-visible:ring-ring",
          "[&_svg]:transition-transform [&_svg]:duration-200",
          "data-[state=on]:[&_svg]:scale-110",
        ],
        ghost: [
          "bg-transparent",
          "hover:bg-muted hover:text-primary",
          "data-[state=on]:bg-primary/10 data-[state=on]:text-primary data-[state=on]:font-medium",
          "data-[state=on]:hover:bg-primary/20",
          "focus-visible:ring-ring",
          "[&_svg]:transition-transform [&_svg]:duration-200",
          "data-[state=on]:[&_svg]:scale-110",
        ],
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        sm: "h-8 rounded-md gap-1.5 px-3 text-xs has-[>svg]:px-2.5",
        lg: "h-11 rounded-md px-8 text-base gap-3 has-[>svg]:px-5",
        icon: "size-9 hover:[&_svg]:rotate-12 data-[state=on]:[&_svg]:rotate-0",
        "icon-sm": "size-8 hover:[&_svg]:rotate-12 data-[state=on]:[&_svg]:rotate-0",
        "icon-lg": "size-11 hover:[&_svg]:rotate-12 data-[state=on]:[&_svg]:rotate-0",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Toggle({
  className,
  variant,
  size,
  ...props
}: React.ComponentProps<typeof TogglePrimitive.Root> &
  VariantProps<typeof toggleVariants>) {
  return (
    <TogglePrimitive.Root
      data-slot="toggle"
      className={cn(toggleVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Toggle, toggleVariants }
