import * as React from "react";
import * as SwitchPrimitives from "@radix-ui/react-switch";
import { cn } from "@/lib/utils";

interface SwitchProps extends React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root> {
  size?: "sm" | "default" | "lg";
  error?: boolean;
}

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitives.Root>,
  SwitchProps
>(({ className, size = "default", error, ...props }, ref) => {
  const sizeClasses = {
    sm: {
      root: "h-4 w-7",
      thumb: "h-3 w-3 data-[state=checked]:translate-x-3"
    },
    default: {
      root: "h-5 w-9",
      thumb: "h-4 w-4 data-[state=checked]:translate-x-4"
    },
    lg: {
      root: "h-6 w-11",
      thumb: "h-5 w-5 data-[state=checked]:translate-x-5"
    }
  };

  const baseStyles = [
    // Base styles
    "peer inline-flex shrink-0 cursor-pointer items-center rounded-full",
    "border-2",
    
    // Transitions
    "transition-all duration-200",
    
    // Focus styles
    "outline-none",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
    
    // Size
    sizeClasses[size].root,
  ];

  const normalStyles = [
    "border-transparent",
    "data-[state=checked]:bg-primary data-[state=unchecked]:bg-muted",
    "hover:data-[state=unchecked]:bg-muted/80",
    "focus-visible:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-brand-500))]"
  ];

  const errorStyles = [
    "border-transparent", 
    "data-[state=checked]:bg-destructive data-[state=unchecked]:bg-destructive/20",
    "hover:data-[state=unchecked]:bg-destructive/30",
    "focus-visible:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-error-600))]"
  ];

  return (
    <SwitchPrimitives.Root
      aria-invalid={error}
      className={cn(
        baseStyles,
        error ? errorStyles : normalStyles,
        className
      )}
      {...props}
      ref={ref}
    >
      <SwitchPrimitives.Thumb
        className={cn(
          "pointer-events-none block rounded-full bg-background shadow-sm ring-0",
          "transition-transform duration-200",
          "data-[state=unchecked]:translate-x-0",
          sizeClasses[size].thumb
        )}
      />
    </SwitchPrimitives.Root>
  );
});

Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };