"use client"

import * as React from "react"
import * as TabsPrimitive from "@radix-ui/react-tabs"

import { cn } from "./lib/utils"

const Tabs = TabsPrimitive.Root

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, children, ...props }, ref) => {
  const [underlineStyle, setUnderlineStyle] = React.useState({ left: 0, width: 0 })
  const listRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    const updatePosition = () => {
      const list = listRef.current
      if (!list) return

      const activeTrigger = list.querySelector('[data-state="active"]') as HTMLElement
      if (activeTrigger) {
        const fullWidth = activeTrigger.offsetWidth
        const underlineWidth = fullWidth * 0.6
        const leftOffset = activeTrigger.offsetLeft + (fullWidth - underlineWidth) / 2

        setUnderlineStyle({
          left: leftOffset,
          width: underlineWidth
        })
      }
    }

    // Update immediately and on resize
    setTimeout(updatePosition, 10)
    const observer = new MutationObserver(updatePosition)
    if (listRef.current) {
      observer.observe(listRef.current, {
        attributes: true,
        subtree: true,
        attributeFilter: ['data-state']
      })
    }

    window.addEventListener('resize', updatePosition)
    return () => {
      observer.disconnect()
      window.removeEventListener('resize', updatePosition)
    }
  }, [children])

  return (
    <TabsPrimitive.List
      ref={(node) => {
        listRef.current = node
        if (typeof ref === 'function') ref(node)
        else if (ref) ref.current = node
      }}
      className={cn(
        "inline-flex gap-1 relative items-center justify-center text-muted-foreground",
        className
      )}
      {...props}
    >
      {children}
      {underlineStyle.width > 0 && (
        <div
          className="absolute bottom-1 h-[2px] bg-primary transition-all duration-300 ease-out"
          style={{
            left: `${underlineStyle.left}px`,
            width: `${underlineStyle.width}px`,
          }}
        />
      )}
    </TabsPrimitive.List>
  )
})
TabsList.displayName = TabsPrimitive.List.displayName

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Trigger
    ref={ref}
    className={cn(
      "relative inline-flex items-center justify-center whitespace-nowrap rounded-sm px-6 py-2.5 text-base font-medium transition-all duration-200",
      "hover:bg-foreground/[0.03] active:bg-foreground/[0.05]",
      "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
      "disabled:pointer-events-none disabled:opacity-50",
      "data-[state=active]:text-foreground",
      "data-[state=inactive]:text-muted-foreground data-[state=inactive]:hover:text-foreground",
      className
    )}
    {...props}
  />
))
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content
    ref={ref}
    className={cn(
      "mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
      className
    )}
    {...props}
  />
))
TabsContent.displayName = TabsPrimitive.Content.displayName

export { Tabs, TabsList, TabsTrigger, TabsContent }
