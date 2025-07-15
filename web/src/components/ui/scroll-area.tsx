"use client"

import * as React from "react"
import * as ScrollAreaPrimitive from "@radix-ui/react-scroll-area"

import { cn } from "@/lib/utils"

type ScrollAreaProps = React.ComponentProps<typeof ScrollAreaPrimitive.Root> & {
  scrollbarVariant?: "default" | "minimal" | "thick" | "rounded"
  hideScrollbar?: boolean
  fadeScrollbar?: boolean
}

function ScrollArea({
  className,
  children,
  scrollbarVariant = "default",
  hideScrollbar = false,
  fadeScrollbar = false,
  ...props
}: ScrollAreaProps) {
  return (
    <ScrollAreaPrimitive.Root
      data-slot="scroll-area"
      className={cn("relative overflow-hidden transition-all duration-200 ease-in-out", className)}
      {...props}
    >
      <ScrollAreaPrimitive.Viewport
        data-slot="scroll-area-viewport"
        className="focus-visible:ring-ring/50 size-full rounded-[inherit] transition-all duration-200 ease-in-out outline-none focus-visible:ring-[3px] focus-visible:outline-1"
      >
        {children}
      </ScrollAreaPrimitive.Viewport>
      {!hideScrollbar && (
        <>
          <ScrollBar variant={scrollbarVariant} fade={fadeScrollbar} />
          <ScrollBar orientation="horizontal" variant={scrollbarVariant} fade={fadeScrollbar} />
        </>
      )}
      <ScrollAreaPrimitive.Corner />
    </ScrollAreaPrimitive.Root>
  )
}

type ScrollBarProps = React.ComponentProps<typeof ScrollAreaPrimitive.ScrollAreaScrollbar> & {
  variant?: "default" | "minimal" | "thick" | "rounded"
  fade?: boolean
}

function ScrollBar({
  className,
  orientation = "vertical",
  variant = "default",
  fade = false,
  ...props
}: ScrollBarProps) {
  const scrollbarVariants = {
    default: {
      bar: "w-2.5 h-2.5",
      thumb: "bg-border hover:bg-border/80 transition-colors duration-200",
    },
    minimal: {
      bar: "w-1.5 h-1.5",
      thumb: "bg-muted-foreground/30 hover:bg-muted-foreground/50 transition-colors duration-200",
    },
    thick: {
      bar: "w-3 h-3",
      thumb: "bg-primary/20 hover:bg-primary/30 transition-colors duration-200",
    },
    rounded: {
      bar: "w-2 h-2",
      thumb: "bg-accent/50 hover:bg-accent/70 transition-colors duration-200 rounded-full",
    },
  }

  const variantStyles = scrollbarVariants[variant]

  return (
    <ScrollAreaPrimitive.ScrollAreaScrollbar
      data-slot="scroll-area-scrollbar"
      orientation={orientation}
      className={cn(
        "flex touch-none p-px transition-all duration-200 select-none",
        orientation === "vertical" && `h-full ${variantStyles.bar.split(' ')[0]} border-l border-l-transparent`,
        orientation === "horizontal" && `${variantStyles.bar.split(' ')[1]} w-full flex-col border-t border-t-transparent`,
        fade && "opacity-0 hover:opacity-100 data-[state=visible]:opacity-100 transition-opacity duration-300",
        className
      )}
      {...props}
    >
      <ScrollAreaPrimitive.ScrollAreaThumb
        data-slot="scroll-area-thumb"
        className={cn(
          "relative flex-1 transition-all duration-200",
          variantStyles.thumb,
          variant === "rounded" ? "rounded-full" : "rounded-sm"
        )}
      />
    </ScrollAreaPrimitive.ScrollAreaScrollbar>
  )
}

export { ScrollArea, ScrollBar }
