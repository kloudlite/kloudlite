"use client"

import * as React from "react"
import { cn } from "@/lib/utils"
import { XMarkIcon } from "@heroicons/react/24/outline"
import { Dialog, DialogPanel, DialogBackdrop, TransitionChild } from "@headlessui/react"
import { cva, type VariantProps } from "class-variance-authority"

const sidebarVariants = cva(
  "flex flex-col bg-card border-r border-border",
  {
    variants: {
      variant: {
        default: "bg-card",
        subtle: "bg-background",
        ghost: "bg-transparent border-transparent",
      },
      size: {
        default: "w-72 max-w-xs",
        sm: "w-64 max-w-[16rem]",
        lg: "w-80 max-w-sm",
        xl: "w-96 max-w-md",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface SidebarProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof sidebarVariants> {
  children?: React.ReactNode
  open?: boolean
  onOpenChange?: (open: boolean) => void
  showMobileCloseButton?: boolean
  mobileBreakpoint?: "sm" | "md" | "lg" | "xl"
}

export interface SidebarContextValue {
  open: boolean
  onOpenChange: (open: boolean) => void
  isMobile: boolean
}

const SidebarContext = React.createContext<SidebarContextValue | undefined>(undefined)

export function useSidebar() {
  const context = React.useContext(SidebarContext)
  if (!context) {
    throw new Error("useSidebar must be used within a Sidebar")
  }
  return context
}

export function Sidebar({
  className,
  variant,
  size,
  children,
  open = false,
  onOpenChange = () => {},
  showMobileCloseButton = true,
  mobileBreakpoint = "lg",
  ...props
}: SidebarProps) {
  const [isMobile, setIsMobile] = React.useState(false)

  React.useEffect(() => {
    const checkMobile = () => {
      const breakpoints = {
        sm: 640,
        md: 768,
        lg: 1024,
        xl: 1280,
      }
      setIsMobile(window.innerWidth < breakpoints[mobileBreakpoint])
    }

    checkMobile()
    window.addEventListener("resize", checkMobile)
    return () => window.removeEventListener("resize", checkMobile)
  }, [mobileBreakpoint])

  const contextValue = React.useMemo(
    () => ({ open, onOpenChange, isMobile }),
    [open, onOpenChange, isMobile]
  )

  if (isMobile) {
    return (
      <SidebarContext.Provider value={contextValue}>
        <Dialog open={open} onClose={onOpenChange} className="relative z-50">
          <DialogBackdrop
            transition
            className="fixed inset-0 bg-background/80 transition-opacity duration-300 ease-linear data-[closed]:opacity-0"
          />
          <div className="fixed inset-0 flex">
            <DialogPanel
              transition
              className={cn(
                "relative mr-16 flex transform transition duration-300 ease-in-out data-[closed]:-translate-x-full",
                sidebarVariants({ variant, size }),
                className
              )}
              {...props}
            >
              {showMobileCloseButton && (
                <TransitionChild>
                  <div className="absolute top-0 left-full flex w-16 justify-center pt-5 duration-300 ease-in-out data-[closed]:opacity-0">
                    <button type="button" onClick={() => onOpenChange(false)} className="-m-2.5 p-2.5">
                      <span className="sr-only">Close sidebar</span>
                      <XMarkIcon aria-hidden="true" className="size-6 text-foreground" />
                    </button>
                  </div>
                </TransitionChild>
              )}
              {children}
            </DialogPanel>
          </div>
        </Dialog>
      </SidebarContext.Provider>
    )
  }

  return (
    <SidebarContext.Provider value={contextValue}>
      <div
        className={cn(
          "hidden fixed inset-y-0 z-50 flex",
          mobileBreakpoint === "sm" && "sm:flex",
          mobileBreakpoint === "md" && "md:flex",
          mobileBreakpoint === "lg" && "lg:flex",
          mobileBreakpoint === "xl" && "xl:flex",
          sidebarVariants({ variant, size }),
          className
        )}
        {...props}
      >
        {children}
      </div>
    </SidebarContext.Provider>
  )
}

export interface SidebarHeaderProps extends React.HTMLAttributes<HTMLDivElement> {}

export function SidebarHeader({ className, ...props }: SidebarHeaderProps) {
  return (
    <div className={cn("flex h-16 shrink-0 items-center", className)} {...props} />
  )
}

export interface SidebarContentProps extends React.HTMLAttributes<HTMLElement> {}

export function SidebarContent({ className, ...props }: SidebarContentProps) {
  return (
    <nav className={cn("flex flex-1 flex-col", className)} {...props} />
  )
}

export interface SidebarGroupProps extends React.HTMLAttributes<HTMLDivElement> {
  label?: string
}

export function SidebarGroup({ label, className, children, ...props }: SidebarGroupProps) {
  return (
    <div className={cn("", className)} {...props}>
      {label && (
        <div className="text-xs/6 font-semibold text-muted-foreground mb-2">{label}</div>
      )}
      <ul role="list" className="-mx-2 space-y-1">
        {children}
      </ul>
    </div>
  )
}

export interface SidebarItemProps extends React.HTMLAttributes<HTMLLIElement> {}

export function SidebarItem({ className, ...props }: SidebarItemProps) {
  return <li className={cn("", className)} {...props} />
}

const sidebarLinkVariants = cva(
  "group flex gap-x-3 rounded-md p-2 text-sm/6 font-medium transition-colors duration-200",
  {
    variants: {
      variant: {
        default: "text-muted-foreground hover:bg-muted hover:text-foreground",
        active: "bg-primary/10 dark:bg-primary/20 text-primary relative before:absolute before:-left-2 before:top-0 before:h-full before:w-1 before:bg-primary",
        ghost: "text-muted-foreground hover:text-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

export interface SidebarLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement>, VariantProps<typeof sidebarLinkVariants> {
  active?: boolean
  icon?: React.ComponentType<{ className?: string }>
}

export const SidebarLink = React.forwardRef<HTMLAnchorElement, SidebarLinkProps>(
  ({ className, variant, active, icon: Icon, children, ...props }, ref) => {
    const linkVariant = active ? "active" : variant
    
    return (
      <a
        ref={ref}
        className={cn(sidebarLinkVariants({ variant: linkVariant }), className)}
        {...props}
      >
        {Icon && (
          <Icon
            aria-hidden="true"
            className={cn(
              "size-6 shrink-0",
              active ? "text-blue-600" : "text-muted-foreground group-hover:text-foreground"
            )}
          />
        )}
        {children}
      </a>
    )
  }
)

SidebarLink.displayName = "SidebarLink"

export interface SidebarFooterProps extends React.HTMLAttributes<HTMLDivElement> {}

export function SidebarFooter({ className, ...props }: SidebarFooterProps) {
  return (
    <div className={cn("mt-auto", className)} {...props} />
  )
}

export interface SidebarToggleProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: React.ComponentType<{ className?: string }>
}

export const SidebarToggle = React.forwardRef<HTMLButtonElement, SidebarToggleProps>(
  ({ className, icon: Icon, children, ...props }, ref) => {
    const { onOpenChange } = useSidebar()
    
    return (
      <button
        ref={ref}
        type="button"
        className={cn(
          "-m-2.5 p-2.5 text-muted-foreground hover:text-foreground",
          className
        )}
        onClick={() => onOpenChange(true)}
        {...props}
      >
        <span className="sr-only">{children || "Open sidebar"}</span>
        {Icon && <Icon className="h-6 w-6" />}
      </button>
    )
  }
)

SidebarToggle.displayName = "SidebarToggle"

// Standalone toggle button that can be used outside of Sidebar context
export interface SidebarTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: React.ComponentType<{ className?: string }>
  onOpen: () => void
}

export const SidebarTrigger = React.forwardRef<HTMLButtonElement, SidebarTriggerProps>(
  ({ className, icon: Icon, children, onOpen, ...props }, ref) => {
    return (
      <button
        ref={ref}
        type="button"
        className={cn(
          "-m-2.5 p-2.5 text-muted-foreground hover:text-foreground",
          className
        )}
        onClick={onOpen}
        {...props}
      >
        <span className="sr-only">{children || "Open sidebar"}</span>
        {Icon && <Icon className="h-6 w-6" />}
      </button>
    )
  }
)

SidebarTrigger.displayName = "SidebarTrigger"

export interface SidebarSeparatorProps extends React.HTMLAttributes<HTMLDivElement> {}

export function SidebarSeparator({ className, ...props }: SidebarSeparatorProps) {
  return (
    <div 
      className={cn("h-px bg-border", className)} 
      {...props} 
    />
  )
}