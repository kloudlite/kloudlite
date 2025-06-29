import * as React from 'react';
import { type LucideIcon } from 'lucide-react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const iconVariants = cva('', {
  variants: {
    size: {
      xs: 'h-3 w-3',    // 12px
      sm: 'h-3.5 w-3.5', // 14px
      base: 'h-4 w-4',   // 16px
      lg: 'h-5 w-5',     // 20px
      xl: 'h-6 w-6',     // 24px
      '2xl': 'h-8 w-8', // 32px
      '3xl': 'h-10 w-10', // 40px
    },
    color: {
      default: 'text-current',
      muted: 'text-muted-foreground',
      primary: 'text-primary',
      secondary: 'text-secondary',
      success: 'text-green-600 dark:text-green-400',
      error: 'text-red-600 dark:text-red-400',
      warning: 'text-yellow-600 dark:text-yellow-400',
      info: 'text-blue-600 dark:text-blue-400',
    },
  },
  defaultVariants: {
    size: 'base',
    color: 'default',
  },
});

export interface IconProps
  extends React.HTMLAttributes<SVGElement>,
    VariantProps<typeof iconVariants> {
  icon: LucideIcon;
  label?: string;
}

const Icon = React.forwardRef<SVGElement, IconProps>(
  ({ icon: IconComponent, size, color, label, className, ...props }, ref) => {
    return (
      <IconComponent
        ref={ref}
        className={cn(iconVariants({ size, color }), className)}
        aria-label={label}
        aria-hidden={!label}
        {...props}
      />
    );
  }
);

Icon.displayName = 'Icon';

export { Icon, iconVariants };