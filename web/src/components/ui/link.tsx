import * as React from "react"
import NextLink from "next/link"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const linkVariants = cva(
  "inline-flex items-center justify-center gap-1 transition-all duration-200 outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background focus-visible:rounded-sm",
  {
    variants: {
      variant: {
        default:
          "text-primary hover:text-primary-hover active:text-primary-active underline-offset-4 hover:underline focus-visible:ring-primary",
        muted:
          "text-muted-foreground hover:text-foreground focus-visible:ring-ring",
        destructive:
          "text-destructive hover:text-destructive-hover active:text-destructive-active underline-offset-4 hover:underline focus-visible:ring-destructive",
        ghost:
          "hover:text-primary active:text-primary-active focus-visible:ring-ring",
      },
      size: {
        default: "text-base",
        sm: "text-sm",
        lg: "text-lg",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface LinkProps
  extends React.ComponentPropsWithoutRef<typeof NextLink>,
    VariantProps<typeof linkVariants> {
  external?: boolean
}

const Link = React.forwardRef<HTMLAnchorElement, LinkProps>(
  ({ className, variant, size, external, ...props }, ref) => {
    const Comp = external ? "a" : NextLink
    
    return (
      <Comp
        className={cn(linkVariants({ variant, size, className }))}
        ref={ref}
        {...(external ? { target: "_blank", rel: "noopener noreferrer" } : {})}
        {...props}
      />
    )
  }
)
Link.displayName = "Link"

export { Link, linkVariants }