import * as React from "react";
import { cn } from "@/lib/utils";

interface SeparatorProps extends React.HTMLAttributes<HTMLDivElement> {
  orientation?: "horizontal" | "vertical";
  decorative?: boolean;
}

const Separator = React.forwardRef<HTMLDivElement, SeparatorProps>(
  ({ className, orientation = "horizontal", decorative = true, ...props }, ref) => {
    const isHorizontal = orientation === "horizontal";
    
    return (
      <div
        ref={ref}
        role={decorative ? "presentation" : "separator"}
        aria-orientation={orientation}
        className={cn(
          "shrink-0 bg-border",
          isHorizontal ? "h-[1px] w-full" : "h-full w-[1px]",
          className
        )}
        {...props}
      />
    );
  }
);

Separator.displayName = "Separator";

export { Separator };