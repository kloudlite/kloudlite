"use client"

import * as React from "react"
import { OTPInput, OTPInputContext } from "input-otp"
import { MinusIcon } from "lucide-react"

import { cn } from "@/lib/utils"

// Add global style to hide overflow dots
if (typeof window !== 'undefined' && !document.getElementById('input-otp-overflow-hide')) {
  const style = document.createElement('style')
  style.id = 'input-otp-overflow-hide'
  style.textContent = `
    /* Hide overflow dots in InputOTP */
    .otp-input-container > span,
    .otp-input-container > :last-child:not(div),
    [data-input-otp-container] > span,
    [data-input-otp-container] > :last-child:not(div) {
      display: none !important;
    }
    /* Also hide any text nodes that might contain dots */
    .otp-input-container::after,
    [data-input-otp-container]::after {
      content: none !important;
    }
  `
  document.head.appendChild(style)
}

function InputOTP({
  className,
  containerClassName,
  ...props
}: React.ComponentProps<typeof OTPInput> & {
  containerClassName?: string
}) {
  return (
    <OTPInput
      data-slot="input-otp"
      containerClassName={cn(
        "otp-input-container flex items-center gap-2 has-disabled:opacity-50",
        // Prevent text selection on the container
        "select-none",
        // Hide overflow indicators with multiple approaches
        "[&>span]:hidden",
        "[&>*:last-child:not(div)]:hidden",
        containerClassName
      )}
      className={cn("disabled:cursor-not-allowed", className)}
      {...props}
    />
  )
}

function InputOTPGroup({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="input-otp-group"
      className={cn(
        "flex items-center group",
        // When multiple slots in this group are active, hide individual rings
        "[&:has([data-active=true]):has([data-active=true]_~_[data-active=true])_[data-active=true]]:ring-0",
        "[&:has([data-active=true]):has([data-active=true]_~_[data-active=true])_[data-active=true]]:ring-offset-0",
        "[&:has([data-active=true]):has([data-active=true]_~_[data-active=true])_[data-active=true]]:scale-100",
        className
      )}
      {...props}
    />
  )
}

function InputOTPSlot({
  index,
  className,
  ...props
}: React.ComponentProps<"div"> & {
  index: number
}) {
  const inputOTPContext = React.useContext(OTPInputContext)
  const { char, hasFakeCaret, isActive } = inputOTPContext?.slots[index] ?? {}
  
  // Check if multiple slots are active (select all scenario)
  const multipleActive = React.useMemo(() => {
    if (!inputOTPContext?.slots) return false
    const activeCount = inputOTPContext.slots.filter(slot => slot.isActive).length
    return activeCount > 1
  }, [inputOTPContext?.slots])
  
  // Check if there's an error (user typed more than maxLength)
  const hasError = React.useMemo(() => {
    if (!inputOTPContext) return false
    const value = props['aria-invalid'] === 'true'
    return value
  }, [inputOTPContext, props])

  return (
    <div
      data-slot="input-otp-slot"
      data-active={isActive}
      className={cn(
        // Base styles
        "relative flex h-10 w-10 items-center justify-center text-sm outline-none",
        // Border and rounded corners for connected design
        "border-y border-r first:rounded-l-md first:border-l last:rounded-r-md",
        // Enhanced transitions
        "transition-all duration-200 ease-in-out",
        // Default state
        "border-input bg-background",
        // Hover state - subtle scale and border color change
        "hover:scale-[1.02] hover:border-muted-foreground/50",
        // Active/focused state with enhanced visual feedback
        "data-[active=true]:border-primary data-[active=true]:bg-primary/5",
        // Special handling for focus ring with connected design - only show ring if truly focused
        !multipleActive && "data-[active=true]:ring-2 data-[active=true]:ring-primary data-[active=true]:ring-offset-2 data-[active=true]:ring-offset-background",
        !multipleActive && "data-[active=true]:scale-105",
        // Fix border visibility by adding left border when active
        "data-[active=true]:border-l data-[active=true]:border-l-primary",
        // Match focus ring shape to slot shape
        "data-[active=true]:first:rounded-l-md data-[active=true]:last:rounded-r-md",
        "data-[active=true]:[&:not(:first-child):not(:last-child)]:rounded-none",
        // Dark mode active state
        "dark:data-[active=true]:bg-primary/10",
        // Filled state (when digit is entered)
        char && "font-medium text-foreground",
        // Empty state
        !char && "text-muted-foreground",
        // Disable text selection highlighting
        "select-none",
        // Error states
        "aria-invalid:border-destructive data-[active=true]:aria-invalid:border-destructive",
        "data-[active=true]:aria-invalid:ring-destructive",
        // Z-index for active state - higher to ensure borders are visible
        "data-[active=true]:z-20",
        // Disabled state
        "disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    >
      {char && (
        <span className="animate-in zoom-in-50 duration-200">
          {char}
        </span>
      )}
      {hasFakeCaret && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <div className="animate-caret-blink bg-primary h-5 w-px duration-1000" />
        </div>
      )}
    </div>
  )
}

function InputOTPSeparator({ ...props }: React.ComponentProps<"div">) {
  return (
    <div data-slot="input-otp-separator" role="separator" {...props}>
      <MinusIcon />
    </div>
  )
}

export { InputOTP, InputOTPGroup, InputOTPSlot, InputOTPSeparator }