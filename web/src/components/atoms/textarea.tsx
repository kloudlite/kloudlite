import * as React from "react"

import { cn } from "@/lib/utils"

interface TextareaProps extends React.ComponentProps<"textarea"> {
  error?: boolean
}

function Textarea({ className, error, ...props }: TextareaProps) {
  const baseStyles = [
    // Base styles
    "flex min-h-[80px] w-full rounded-md",
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
    
    // Resize
    "resize-y",
  ];

  const normalStyles = [
    "border-form-border",
    "hover:border-form-border-hover",
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
    <textarea
      data-slot="textarea"
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

export { Textarea }
