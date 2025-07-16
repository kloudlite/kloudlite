"use client"

import * as React from "react"
import * as ProgressPrimitive from "@radix-ui/react-progress"

import { cn } from "@/lib/utils"

interface ProgressProps extends React.ComponentProps<typeof ProgressPrimitive.Root> {
  indicatorClassName?: string
  showAnimation?: boolean
}

function Progress({
  className,
  value,
  indicatorClassName,
  showAnimation = true,
  ...props
}: ProgressProps) {
  const [animatedValue, setAnimatedValue] = React.useState(0)

  React.useEffect(() => {
    if (showAnimation) {
      const timeout = setTimeout(() => {
        setAnimatedValue(value || 0)
      }, 100)
      return () => clearTimeout(timeout)
    } else {
      setAnimatedValue(value || 0)
    }
  }, [value, showAnimation])

  return (
    <ProgressPrimitive.Root
      data-slot="progress"
      className={cn(
        "relative h-2 w-full overflow-hidden rounded-none bg-secondary/20",
        "group hover:bg-secondary/30 transition-colors duration-300",
        className
      )}
      {...props}
    >
      <ProgressPrimitive.Indicator
        data-slot="progress-indicator"
        className={cn(
          "h-full w-full flex-1",
          "bg-primary",
          "transition-all duration-500 ease-out",
          "relative overflow-hidden",
          "after:absolute after:inset-0",
          "after:bg-gradient-to-r after:from-transparent after:via-white/20 after:to-transparent",
          "after:translate-x-[-200%] group-hover:after:translate-x-[200%]",
          "after:transition-transform after:duration-1000",
          indicatorClassName
        )}
        style={{ 
          transform: `translateX(-${100 - animatedValue}%)`,
        }}
      />
    </ProgressPrimitive.Root>
  )
}

export { Progress }
