"use client"

import * as React from "react"
import * as TabsPrimitive from "@radix-ui/react-tabs"

import { cn } from "@/lib/utils"

type TabsProps = React.ComponentProps<typeof TabsPrimitive.Root> & {
  animateContent?: boolean
}

const TabsContext = React.createContext<{ animateContent?: boolean }>({})

function Tabs({
  className,
  animateContent = false,
  ...props
}: TabsProps) {
  return (
    <TabsContext.Provider value={{ animateContent }}>
      <TabsPrimitive.Root
        data-slot="tabs"
        className={cn("flex flex-col gap-3", className)}
        {...props}
      />
    </TabsContext.Provider>
  )
}

function TabsList({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.List>) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      className={cn(
        "bg-muted text-muted-foreground inline-flex h-10 w-fit items-center justify-center rounded-lg p-1",
        "transition-all duration-200 ease-in-out",
        className
      )}
      {...props}
    />
  )
}

function TabsTrigger({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Trigger>) {
  return (
    <TabsPrimitive.Trigger
      data-slot="tabs-trigger"
      className={cn(
        "inline-flex h-8 flex-1 items-center justify-center gap-2 rounded-md border border-transparent px-3 py-1.5",
        "text-sm font-medium whitespace-nowrap transition-all duration-200 ease-in-out",
        "text-muted-foreground hover:text-foreground",
        "hover:bg-background/50 active:bg-background/80",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        "disabled:pointer-events-none disabled:opacity-50",
        "data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm",
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    />
  )
}

function TabsContent({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Content>) {
  const { animateContent } = React.useContext(TabsContext)
  
  return (
    <TabsPrimitive.Content
      data-slot="tabs-content"
      className={cn(
        "flex-1 mt-2",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        animateContent && "animate-in fade-in-0 zoom-in-95 duration-200",
        className
      )}
      tabIndex={-1}
      {...props}
    />
  )
}

export { Tabs, TabsList, TabsTrigger, TabsContent }
