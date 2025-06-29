import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const headingVariants = cva('font-sans', {
  variants: {
    level: {
      1: cn('text-5xl font-bold tracking-tight'),
      2: cn('text-4xl font-bold tracking-tight'),
      3: cn('text-3xl font-semibold tracking-tight'),
      4: cn('text-2xl font-semibold tracking-tight'),
      5: cn('text-xl font-semibold'),
      6: cn('text-lg font-semibold'),
    },
    size: {
      xs: 'text-xs',
      sm: 'text-sm',
      base: 'text-base',
      lg: 'text-lg',
      xl: 'text-xl',
      '2xl': 'text-2xl',
      '3xl': 'text-3xl',
      '4xl': 'text-4xl',
      '5xl': 'text-5xl',
      '6xl': 'text-6xl',
    },
    weight: {
      light: 'font-light',
      normal: 'font-normal',
      medium: 'font-medium',
      semibold: 'font-semibold',
      bold: 'font-bold',
    },
    color: {
      default: 'text-foreground',
      secondary: 'text-foreground/70',
      muted: 'text-muted-foreground',
      inverse: 'text-background',
      success: 'text-success',
      error: 'text-error',
      warning: 'text-warning',
      info: 'text-info',
      primary: 'text-primary',
      inherit: '',
    },
    align: {
      left: 'text-left',
      center: 'text-center',
      right: 'text-right',
      justify: 'text-justify',
    },
    tracking: {
      tighter: 'tracking-tighter',
      tight: 'tracking-tight',
      normal: 'tracking-normal',
      wide: 'tracking-wide',
      wider: 'tracking-wider',
      widest: 'tracking-widest',
    },
    leading: {
      none: 'leading-none',
      tight: 'leading-tight',
      snug: 'leading-snug',
      normal: 'leading-normal',
      relaxed: 'leading-relaxed',
      loose: 'leading-loose',
    },
  },
  defaultVariants: {
    level: 1,
    weight: 'semibold',
    color: 'default',
    align: 'left',
    tracking: 'tight',
    leading: 'tight',
  },
});

export interface HeadingProps
  extends React.HTMLAttributes<HTMLHeadingElement>,
    VariantProps<typeof headingVariants> {
  as?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';
}

const Heading = React.forwardRef<HTMLHeadingElement, HeadingProps>(
  ({ 
    className, 
    level = 1, 
    size, 
    weight, 
    color, 
    align, 
    tracking, 
    leading,
    as, 
    children, 
    ...props 
  }, ref) => {
    const Component = as || (`h${level}` as keyof JSX.IntrinsicElements);
    
    // If size is provided, use it with other variants
    // Otherwise, use level which includes predefined size, weight, and tracking
    const variantProps = size 
      ? { size, weight, color, align, tracking, leading }
      : { level, color, align };
    
    return (
      <Component
        ref={ref}
        className={cn(
          headingVariants(variantProps), 
          className
        )}
        {...props}
      >
        {children}
      </Component>
    );
  }
);

Heading.displayName = 'Heading';

export { Heading, headingVariants };