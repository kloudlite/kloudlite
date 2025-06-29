import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';
import { Icon, type IconProps } from './icon';

const iconBoxVariants = cva(
  'flex items-center justify-center rounded-lg',
  {
    variants: {
      size: {
        xs: 'h-6 w-6',
        sm: 'h-8 w-8',
        md: 'h-10 w-10',
        default: 'h-10 w-10',
        lg: 'h-12 w-12',
        xl: 'h-14 w-14',
      },
      color: {
        // Neutral colors
        gray: 'bg-gray-100 dark:bg-gray-800',
        slate: 'bg-slate-100 dark:bg-slate-800',
        muted: 'bg-gray-50 dark:bg-gray-900',
        
        // Status colors - Professional palette
        green: 'bg-emerald-50 dark:bg-emerald-950/30',
        emerald: 'bg-emerald-50 dark:bg-emerald-950/30',
        blue: 'bg-sky-50 dark:bg-sky-950/30',
        sky: 'bg-sky-50 dark:bg-sky-950/30',
        red: 'bg-rose-50 dark:bg-rose-950/30',
        rose: 'bg-rose-50 dark:bg-rose-950/30',
        yellow: 'bg-amber-50 dark:bg-amber-950/30',
        amber: 'bg-amber-50 dark:bg-amber-950/30',
        
        // Accent colors
        purple: 'bg-purple-50 dark:bg-purple-950/30',
        violet: 'bg-violet-50 dark:bg-violet-950/30',
        indigo: 'bg-indigo-50 dark:bg-indigo-950/30',
        
        // Additional colors
        orange: 'bg-orange-50 dark:bg-orange-950/30',
        pink: 'bg-pink-50 dark:bg-pink-950/30',
        
        // Special
        gradient: 'bg-gradient-to-br from-blue-500 to-blue-700',
      },
    },
    defaultVariants: {
      size: 'default',
      color: 'slate',
    },
  }
);

const iconSizeMap = {
  xs: 'sm' as const,
  sm: 'base' as const,
  md: 'lg' as const,
  default: 'lg' as const,
  lg: 'xl' as const,
  xl: '2xl' as const,
};

const iconColorMap = {
  // Neutral colors
  gray: 'text-gray-600 dark:text-gray-400',
  slate: 'text-slate-600 dark:text-slate-400',
  muted: 'text-gray-500 dark:text-gray-400',
  
  // Status colors - Professional palette
  green: 'text-emerald-700 dark:text-emerald-400',
  emerald: 'text-emerald-700 dark:text-emerald-400',
  blue: 'text-sky-700 dark:text-sky-400',
  sky: 'text-sky-700 dark:text-sky-400',
  red: 'text-rose-700 dark:text-rose-400',
  rose: 'text-rose-700 dark:text-rose-400',
  yellow: 'text-amber-700 dark:text-amber-400',
  amber: 'text-amber-700 dark:text-amber-400',
  
  // Brand colors
  purple: 'text-violet-700 dark:text-violet-400',
  violet: 'text-violet-700 dark:text-violet-400',
  indigo: 'text-indigo-700 dark:text-indigo-400',
  
  // Additional colors
  orange: 'text-orange-700 dark:text-orange-400',
  pink: 'text-pink-700 dark:text-pink-400',
  
  // Special
  gradient: 'text-white',
};

export interface IconBoxProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof iconBoxVariants> {
  icon: IconProps['icon'];
  iconClassName?: string;
}

const IconBox = React.forwardRef<HTMLDivElement, IconBoxProps>(
  ({ className, size, color = 'slate', icon, iconClassName, ...props }, ref) => {
    const iconSize = iconSizeMap[size || 'default'];
    const iconColor = iconColorMap[color];
    
    return (
      <div
        ref={ref}
        className={cn(iconBoxVariants({ size, color }), className)}
        {...props}
      >
        <Icon 
          icon={icon} 
          size={iconSize}
          className={cn(iconColor, iconClassName)}
        />
      </div>
    );
  }
);

IconBox.displayName = 'IconBox';

export { IconBox, iconBoxVariants };