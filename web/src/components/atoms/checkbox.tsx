"use client"

import * as React from "react"
import * as CheckboxPrimitive from "@radix-ui/react-checkbox"
import { CheckIcon } from "lucide-react"

import { cn } from "@/lib/utils"

interface CheckboxProps extends React.ComponentProps<typeof CheckboxPrimitive.Root> {
  error?: boolean
}

function Checkbox({
  className,
  error,
  ...props
}: CheckboxProps) {
  const baseStyles = [
    // Base styles
    "peer h-4 w-4 shrink-0 rounded-sm",
    
    // Background
    "bg-background",
    
    // Border
    "border",
    
    // Focus styles
    "outline-none",
    
    // Transitions
    "transition-all duration-200",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
    
    // Checked state
    "data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground data-[state=checked]:border-primary",
  ];

  const normalStyles = [
    "border-form-border",
    "hover:border-form-border-hover",
    "focus-visible:border-form-border-focus",
    "focus-visible:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-brand-500))]"
  ];

  const errorStyles = [
    "border-destructive/70",
    "hover:border-destructive/80",
    "focus-visible:border-destructive",
    "focus-visible:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-error-600))]"
  ];

  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      aria-invalid={error || props['aria-invalid']}
      className={cn(
        baseStyles,
        error ? errorStyles : normalStyles,
        className
      )}
      {...props}
    >
      <CheckboxPrimitive.Indicator
        data-slot="checkbox-indicator"
        className="flex items-center justify-center text-current transition-none"
      >
        <CheckIcon className="size-3.5" />
      </CheckboxPrimitive.Indicator>
    </CheckboxPrimitive.Root>
  )
}

export { Checkbox }
