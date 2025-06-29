"use client";

import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const avatarVariants = cva(
  'inline-flex items-center justify-center font-medium rounded-full overflow-hidden relative',
  {
    variants: {
      size: {
        xs: 'h-6 w-6 text-xs',
        sm: 'h-8 w-8 text-sm',
        md: 'h-10 w-10 text-base',
        default: 'h-10 w-10 text-base',
        lg: 'h-12 w-12 text-lg',
        xl: 'h-14 w-14 text-xl',
      },
      color: {
        slate: 'bg-muted text-muted-foreground',
        blue: 'bg-[rgb(219_234_254)] dark:bg-[rgb(30_58_138)] text-[rgb(29_78_216)] dark:text-[rgb(191_219_254)]',
        green: 'bg-[rgb(209_250_229)] dark:bg-[rgb(6_78_59)] text-[rgb(4_120_87)] dark:text-[rgb(167_243_208)]',
        yellow: 'bg-[rgb(254_243_199)] dark:bg-[rgb(120_53_15)] text-[rgb(180_83_9)] dark:text-[rgb(253_230_138)]',
        red: 'bg-[rgb(255_228_230)] dark:bg-[rgb(127_29_29)] text-[rgb(190_18_60)] dark:text-[rgb(254_202_202)]',
        purple: 'bg-[rgb(224_231_255)] dark:bg-[rgb(55_48_163)] text-[rgb(67_56_202)] dark:text-[rgb(199_210_254)]',
        gradient: 'bg-gradient-to-br from-[rgb(59_130_246)] to-[rgb(6_182_212)] text-white',
        muted: 'bg-secondary text-secondary-foreground',
      },
    },
    defaultVariants: {
      size: 'default',
      color: 'slate',
    },
  }
);

export interface AvatarProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof avatarVariants> {
  src?: string;
  alt?: string;
  fallback?: string;
}

const Avatar = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ className, size, color, src, alt, fallback, children, ...props }, ref) => {
    const [imageError, setImageError] = React.useState(false);
    
    const showFallback = !src || imageError;
    
    return (
      <div
        ref={ref}
        className={cn(avatarVariants({ size, color }), className)}
        {...props}
      >
        {src && !imageError && (
          <img
            src={src}
            alt={alt || ''}
            onError={() => setImageError(true)}
            className="h-full w-full object-cover"
          />
        )}
        {showFallback && (
          <span className="select-none font-semibold">
            {fallback || children}
          </span>
        )}
      </div>
    );
  }
);

Avatar.displayName = 'Avatar';

export { Avatar, avatarVariants };