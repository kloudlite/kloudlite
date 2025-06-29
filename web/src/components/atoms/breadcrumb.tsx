import * as React from 'react';
import { ChevronRight, LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';
import Link from 'next/link';
import { cva, type VariantProps } from 'class-variance-authority';

const breadcrumbVariants = cva(
  "flex items-center gap-2",
  {
    variants: {
      size: {
        sm: "text-xs",
        default: "text-sm",
        lg: "text-base",
      },
    },
    defaultVariants: {
      size: "default",
    },
  }
);

const breadcrumbItemVariants = cva(
  "inline-flex items-center gap-1.5 transition-colors",
  {
    variants: {
      size: {
        sm: "[&>svg]:h-3 [&>svg]:w-3",
        default: "[&>svg]:h-4 [&>svg]:w-4",
        lg: "[&>svg]:h-5 [&>svg]:w-5",
      },
    },
    defaultVariants: {
      size: "default",
    },
  }
);

const breadcrumbSeparatorVariants = cva(
  "text-muted-foreground/70",
  {
    variants: {
      size: {
        sm: "[&>svg]:h-3 [&>svg]:w-3",
        default: "[&>svg]:h-4 [&>svg]:w-4", 
        lg: "[&>svg]:h-5 [&>svg]:w-5",
      },
    },
    defaultVariants: {
      size: "default",
    },
  }
);

export interface BreadcrumbItem {
  label: string;
  href?: string;
  icon?: LucideIcon;
}

export interface BreadcrumbProps extends React.HTMLAttributes<HTMLDivElement>, 
  VariantProps<typeof breadcrumbVariants> {
  items: BreadcrumbItem[];
  separator?: React.ReactNode;
}

const Breadcrumb = React.forwardRef<HTMLDivElement, BreadcrumbProps>(
  ({ className, items, separator, size, ...props }, ref) => {
    const defaultSeparator = <ChevronRight />;
    
    return (
      <div 
        ref={ref} 
        className={cn(breadcrumbVariants({ size }), "text-muted-foreground", className)}
        {...props}
      >
        {items.map((item, index) => {
          const Icon = item.icon;
          const isLast = index === items.length - 1;
          const content = (
            <>
              {Icon && <Icon />}
              <span>{item.label}</span>
            </>
          );

          return (
            <React.Fragment key={index}>
              {item.href && !isLast ? (
                <Link 
                  href={item.href}
                  className={cn(
                    breadcrumbItemVariants({ size }),
                    "hover:text-foreground"
                  )}
                >
                  {content}
                </Link>
              ) : (
                <span className={cn(
                  breadcrumbItemVariants({ size }),
                  isLast && "text-foreground font-medium"
                )}>
                  {content}
                </span>
              )}
              {!isLast && (
                <span className={breadcrumbSeparatorVariants({ size })}>
                  {separator || defaultSeparator}
                </span>
              )}
            </React.Fragment>
          );
        })}
      </div>
    );
  }
);

Breadcrumb.displayName = 'Breadcrumb';

export { Breadcrumb };