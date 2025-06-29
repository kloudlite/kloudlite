import * as React from 'react';
import { cn } from '@/lib/utils';
import { Text, type TextProps } from './text';

export interface CaptionProps extends Omit<TextProps, 'size' | 'color'> {
  error?: boolean;
}

const Caption = React.forwardRef<HTMLParagraphElement, CaptionProps>(
  ({ className, error, weight = 'normal', ...props }, ref) => {
    return (
      <Text
        ref={ref}
        size="sm"
        color={error ? 'error' : 'muted'}
        weight={weight}
        className={cn('mt-1', className)}
        {...props}
      />
    );
  }
);

Caption.displayName = 'Caption';

export { Caption };