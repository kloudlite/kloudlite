"use client"

import * as React from "react"
import * as LabelPrimitive from "@radix-ui/react-label"

import { cn } from "@/lib/utils"

function Label({
  className,
  ...props
}: React.ComponentProps<typeof LabelPrimitive.Root>) {
  return (
    <LabelPrimitive.Root
      data-slot="label"
      className={cn(
        "flex items-center gap-2 text-sm leading-none font-medium select-none transition-all duration-200 ease-in-out",
        "group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50",
        "peer-disabled:cursor-not-allowed peer-disabled:opacity-50",
        "group-data-[invalid=true]:text-destructive",
        "group-data-[required=true]:after:content-['*'] group-data-[required=true]:after:text-destructive group-data-[required=true]:after:ml-0.5",
        className
      )}
      {...props}
    />
  )
}

export { Label }
