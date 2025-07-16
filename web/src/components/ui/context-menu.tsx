"use client"

import * as React from "react"
import * as ContextMenuPrimitive from "@radix-ui/react-context-menu"
import { CheckIcon, ChevronRightIcon } from "lucide-react"

import { cn } from "@/lib/utils"

function ContextMenu({
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Root>) {
  return <ContextMenuPrimitive.Root data-slot="context-menu" {...props} />
}

const ContextMenuTrigger = React.forwardRef<
  React.ElementRef<typeof ContextMenuPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Trigger>
>(({ className, children, asChild, ...props }, ref) => {
  const [showRipple, setShowRipple] = React.useState(false)
  const [ripplePosition, setRipplePosition] = React.useState({ x: 0, y: 0 })

  const handleContextMenu = (e: React.MouseEvent) => {
    const rect = e.currentTarget.getBoundingClientRect()
    setRipplePosition({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    })
    setShowRipple(true)
    setTimeout(() => setShowRipple(false), 600)
  }

  // If asChild is true, we need to handle the ripple differently
  if (asChild) {
    return (
      <ContextMenuPrimitive.Trigger
        ref={ref}
        asChild
        onContextMenu={handleContextMenu}
        {...props}
      >
        {children}
      </ContextMenuPrimitive.Trigger>
    )
  }

  return (
    <ContextMenuPrimitive.Trigger
      ref={ref}
      className={cn("relative", className)}
      onContextMenu={handleContextMenu}
      {...props}
    >
      {children}
      {showRipple && (
        <span
          className="pointer-events-none absolute z-10 animate-ripple rounded-none bg-primary/20"
          style={{
            left: ripplePosition.x - 20,
            top: ripplePosition.y - 20,
            width: 40,
            height: 40,
          }}
        />
      )}
    </ContextMenuPrimitive.Trigger>
  )
})
ContextMenuTrigger.displayName = ContextMenuPrimitive.Trigger.displayName

function ContextMenuGroup({
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Group>) {
  return (
    <ContextMenuPrimitive.Group data-slot="context-menu-group" {...props} />
  )
}

function ContextMenuPortal({
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Portal>) {
  return (
    <ContextMenuPrimitive.Portal data-slot="context-menu-portal" {...props} />
  )
}

function ContextMenuSub({
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Sub>) {
  return <ContextMenuPrimitive.Sub data-slot="context-menu-sub" {...props} />
}

function ContextMenuRadioGroup({
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.RadioGroup>) {
  return (
    <ContextMenuPrimitive.RadioGroup
      data-slot="context-menu-radio-group"
      {...props}
    />
  )
}

function ContextMenuSubTrigger({
  className,
  inset,
  children,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.SubTrigger> & {
  inset?: boolean
}) {
  return (
    <ContextMenuPrimitive.SubTrigger
      data-slot="context-menu-sub-trigger"
      data-inset={inset}
      className={cn(
        "group flex cursor-default items-center rounded-none px-2 py-1.5 text-sm outline-hidden select-none",
        // Enhanced hover and focus states
        "hover:bg-primary-subtle-hover",
        "focus:bg-primary focus:text-primary-foreground",
        "data-[state=open]:bg-primary data-[state=open]:text-primary-foreground",
        // Smooth transitions
        "transition-all duration-150",
        // Icon styling
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 [&_svg]:text-inherit",
        "data-[inset]:pl-8",
        className
      )}
      {...props}
    >
      {children}
      <ChevronRightIcon className="ml-auto transition-transform duration-200 group-data-[state=open]:rotate-90" />
    </ContextMenuPrimitive.SubTrigger>
  )
}

function ContextMenuSubContent({
  className,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.SubContent>) {
  return (
    <ContextMenuPrimitive.SubContent
      data-slot="context-menu-sub-content"
      className={cn(
        "bg-popover text-popover-foreground z-50 min-w-[8rem] origin-(--radix-context-menu-content-transform-origin) overflow-hidden rounded-none border p-1 shadow-lg",
        // Enhanced animations
        "data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        "data-[state=closed]:zoom-out-90 data-[state=open]:zoom-in-95",
        // Position-aware slide animations for submenus
        "data-[side=bottom]:slide-in-from-top-1 data-[side=left]:slide-in-from-right-1",
        "data-[side=right]:slide-in-from-left-1 data-[side=top]:slide-in-from-bottom-1",
        // Smooth transition
        "transition-all duration-150",
        className
      )}
      {...props}
    />
  )
}

function ContextMenuContent({
  className,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Content>) {
  return (
    <ContextMenuPrimitive.Portal>
      <ContextMenuPrimitive.Content
        data-slot="context-menu-content"
        className={cn(
          "bg-popover text-popover-foreground z-50 min-w-[8rem] min-h-[2rem] origin-(--radix-context-menu-content-transform-origin) overflow-hidden rounded-none border p-1 shadow-md",
          // Enhanced animations with position awareness
          "data-[state=open]:animate-in data-[state=closed]:animate-out",
          "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
          "data-[state=closed]:zoom-out-90 data-[state=open]:zoom-in-95",
          // Position-aware slide animations
          "data-[side=bottom]:slide-in-from-top-1 data-[side=left]:slide-in-from-right-1",
          "data-[side=right]:slide-in-from-left-1 data-[side=top]:slide-in-from-bottom-1",
          // Smooth transition
          "transition-all duration-150",
          className
        )}
        {...props}
      />
    </ContextMenuPrimitive.Portal>
  )
}

function ContextMenuItem({
  className,
  inset,
  variant = "default",
  children,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Item> & {
  inset?: boolean
  variant?: "default" | "destructive"
}) {
  return (
    <ContextMenuPrimitive.Item
      data-slot="context-menu-item"
      data-inset={inset}
      data-variant={variant}
      className={cn(
        "group relative flex w-full cursor-default items-center gap-2 rounded-none py-1.5 pr-2 pl-2 text-sm outline-hidden select-none",
        // Enhanced hover and focus states
        "hover:bg-primary-subtle-hover focus:bg-primary focus:text-primary-foreground",
        // Smooth transitions
        "transition-all duration-150",
        // Disabled state
        "data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
        // Icon handling
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 [&_svg]:text-inherit",
        // Destructive variant
        "data-[variant=destructive]:text-destructive",
        "data-[variant=destructive]:hover:bg-destructive-subtle-hover",
        "data-[variant=destructive]:focus:bg-destructive data-[variant=destructive]:focus:text-destructive-foreground",
        // Inset
        "data-[inset]:pl-8",
        className
      )}
      {...props}
    >
      {children}
    </ContextMenuPrimitive.Item>
  )
}

function ContextMenuCheckboxItem({
  className,
  children,
  checked,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.CheckboxItem>) {
  return (
    <ContextMenuPrimitive.CheckboxItem
      data-slot="context-menu-checkbox-item"
      className={cn(
        "relative flex cursor-default items-center gap-2 rounded-none py-1.5 pr-2 pl-8 text-sm outline-hidden select-none",
        // Enhanced hover and focus states
        "hover:bg-primary-subtle-hover",
        "focus:bg-primary focus:text-primary-foreground",
        // Smooth transitions
        "transition-all duration-150",
        "data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      )}
      checked={checked}
      {...props}
    >
      <span className="pointer-events-none absolute left-2 flex size-3.5 items-center justify-center">
        <ContextMenuPrimitive.ItemIndicator>
          <CheckIcon className="size-4 text-current transition-all duration-200 data-[state=checked]:scale-100 data-[state=unchecked]:scale-0" />
        </ContextMenuPrimitive.ItemIndicator>
      </span>
      {children}
    </ContextMenuPrimitive.CheckboxItem>
  )
}

function ContextMenuRadioItem({
  className,
  children,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.RadioItem>) {
  return (
    <ContextMenuPrimitive.RadioItem
      data-slot="context-menu-radio-item"
      className={cn(
        "relative flex cursor-default items-center gap-2 rounded-none py-1.5 pr-2 pl-8 text-sm outline-hidden select-none",
        // Enhanced hover and focus states
        "hover:bg-primary-subtle-hover",
        "focus:bg-primary focus:text-primary-foreground",
        // Smooth transitions
        "transition-all duration-150",
        "data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    >
      <span className="pointer-events-none absolute left-2 flex size-3.5 items-center justify-center">
        <ContextMenuPrimitive.ItemIndicator>
          <div className="size-2 rounded-none bg-current transition-all duration-200 data-[state=checked]:scale-100 data-[state=unchecked]:scale-0" />
        </ContextMenuPrimitive.ItemIndicator>
      </span>
      {children}
    </ContextMenuPrimitive.RadioItem>
  )
}

function ContextMenuLabel({
  className,
  inset,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Label> & {
  inset?: boolean
}) {
  return (
    <ContextMenuPrimitive.Label
      data-slot="context-menu-label"
      data-inset={inset}
      className={cn(
        "text-muted-foreground px-2 py-1.5 text-xs data-[inset]:pl-8",
        className
      )}
      {...props}
    />
  )
}

function ContextMenuSeparator({
  className,
  ...props
}: React.ComponentProps<typeof ContextMenuPrimitive.Separator>) {
  return (
    <ContextMenuPrimitive.Separator
      data-slot="context-menu-separator"
      className={cn("bg-border -mx-1 my-1 h-px", className)}
      {...props}
    />
  )
}

function ContextMenuShortcut({
  className,
  ...props
}: React.ComponentProps<"span">) {
  return (
    <span
      data-slot="context-menu-shortcut"
      className={cn(
        "ml-auto text-xs tracking-widest text-muted-foreground group-focus:text-current",
        className
      )}
      {...props}
    />
  )
}

export {
  ContextMenu,
  ContextMenuTrigger,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuCheckboxItem,
  ContextMenuRadioItem,
  ContextMenuLabel,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuGroup,
  ContextMenuPortal,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuRadioGroup,
}