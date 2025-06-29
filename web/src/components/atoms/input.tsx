import * as React from "react"

import { cn } from "@/lib/utils"

interface InputProps extends React.ComponentProps<"input"> {
  error?: boolean
}

function Input({ className, type, error, ...props }: InputProps) {
  const baseStyles = [
    // Base styles
    "flex h-10 w-full rounded-md",
    "px-3 py-2",
    "text-sm",
    
    // Background
    "bg-background",
    
    // Border
    "border",
    
    // Typography
    "text-foreground",
    "placeholder:text-muted-foreground",
    
    // Focus styles
    "outline-none",
    
    // Transitions
    "transition-all duration-200",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
    
    // File input
    "file:border-0 file:bg-transparent",
    "file:text-sm file:font-medium",
  ];

  const normalStyles = [
    "border-border",
    "hover:border-border/80",
    "focus:border-form-border-focus",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-brand-500))]"
  ];

  const errorStyles = [
    "border-error",
    "hover:border-error",
    "focus:border-error",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_var(--color-error)]"
  ];

  return (
    <input
      type={type}
      data-slot="input"
      aria-invalid={error || props['aria-invalid']}
      className={cn(
        baseStyles,
        error ? errorStyles : normalStyles,
        className
      )}
      {...props}
    />
  )
}

export { Input }
