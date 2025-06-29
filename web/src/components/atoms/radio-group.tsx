"use client"

import * as React from "react"
import * as RadioGroupPrimitive from "@radix-ui/react-radio-group"
import { CircleIcon } from "lucide-react"

import { cn } from "@/lib/utils"

function RadioGroup({
  className,
  ...props
}: React.ComponentProps<typeof RadioGroupPrimitive.Root>) {
  return (
    <RadioGroupPrimitive.Root
      data-slot="radio-group"
      className={cn("grid gap-3", className)}
      {...props}
    />
  )
}

interface RadioGroupItemProps extends React.ComponentProps<typeof RadioGroupPrimitive.Item> {
  error?: boolean
}

function RadioGroupItem({
  className,
  error,
  ...props
}: RadioGroupItemProps) {
  const baseStyles = [
    // Base styles
    "aspect-square h-4 w-4 rounded-full",
    
    // Background
    "bg-background text-primary",
    
    // Border
    "border",
    
    // Focus styles
    "outline-none",
    
    // Transitions
    "transition-all duration-200",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
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
    <RadioGroupPrimitive.Item
      data-slot="radio-group-item"
      aria-invalid={error || props['aria-invalid']}
      className={cn(
        baseStyles,
        error ? errorStyles : normalStyles,
        className
      )}
      {...props}
    >
      <RadioGroupPrimitive.Indicator
        data-slot="radio-group-indicator"
        className="relative flex items-center justify-center"
      >
        <CircleIcon className="fill-form-indicator absolute top-1/2 left-1/2 size-2.5 -translate-x-1/2 -translate-y-1/2" />
      </RadioGroupPrimitive.Indicator>
    </RadioGroupPrimitive.Item>
  )
}

export { RadioGroup, RadioGroupItem }
