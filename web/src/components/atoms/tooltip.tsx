"use client";

import * as React from "react";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const TooltipProvider = TooltipPrimitive.Provider;
const TooltipRoot = TooltipPrimitive.Root;
const TooltipTrigger = TooltipPrimitive.Trigger;

const tooltipContentVariants = cva(
  "z-50 overflow-hidden rounded-md px-3 py-1.5 text-xs animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
  {
    variants: {
      variant: {
        default: "bg-popover text-popover-foreground border shadow-md",
        secondary: "bg-secondary text-secondary-foreground border",
        destructive: "bg-destructive text-destructive-foreground",
        dark: "bg-foreground text-background shadow-md",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

const TooltipContent = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content> &
    VariantProps<typeof tooltipContentVariants>
>(({ className, variant, sideOffset = 4, ...props }, ref) => (
  <TooltipPrimitive.Content
    ref={ref}
    sideOffset={sideOffset}
    className={cn(tooltipContentVariants({ variant }), className)}
    {...props}
  />
));
TooltipContent.displayName = TooltipPrimitive.Content.displayName;

interface TooltipProps {
  children: React.ReactNode;
  content: React.ReactNode;
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  delayDuration?: number;
  variant?: VariantProps<typeof tooltipContentVariants>["variant"];
  side?: "top" | "right" | "bottom" | "left";
  align?: "start" | "center" | "end";
  className?: string;
}

function Tooltip({
  children,
  content,
  open,
  defaultOpen,
  onOpenChange,
  delayDuration = 200,
  variant = "default",
  side = "top",
  align = "center",
  className,
}: TooltipProps) {
  return (
    <TooltipProvider delayDuration={delayDuration}>
      <TooltipRoot open={open} defaultOpen={defaultOpen} onOpenChange={onOpenChange}>
        <TooltipTrigger asChild>{children}</TooltipTrigger>
        <TooltipContent variant={variant} side={side} align={align} className={className}>
          {content}
        </TooltipContent>
      </TooltipRoot>
    </TooltipProvider>
  );
}

export {
  Tooltip,
  TooltipProvider,
  TooltipRoot,
  TooltipTrigger,
  TooltipContent,
  tooltipContentVariants,
};