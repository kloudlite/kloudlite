import * as React from 'react';
import { cn } from '@/lib/utils';
import { Icon, type IconProps } from './icon';

export interface StatDisplayProps extends React.HTMLAttributes<HTMLDivElement> {
  icon?: IconProps['icon'];
  value: string | number;
  label: string;
  size?: 'sm' | 'default' | 'lg';
}

const StatDisplay = React.forwardRef<HTMLDivElement, StatDisplayProps>(
  ({ className, icon, value, label, size = 'default', ...props }, ref) => {
    const sizeClasses = {
      sm: {
        container: 'text-center',
        value: 'text-lg',
        label: 'text-xs',
        icon: 'h-3.5 w-3.5',
        gap: 'gap-1',
        spacing: 'mb-0.5',
      },
      default: {
        container: 'text-center',
        value: 'text-2xl',
        label: 'text-sm',
        icon: 'h-4 w-4',
        gap: 'gap-2',
        spacing: 'mb-1',
      },
      lg: {
        container: 'text-center',
        value: 'text-3xl',
        label: 'text-base',
        icon: 'h-5 w-5',
        gap: 'gap-2',
        spacing: 'mb-2',
      },
    };

    const config = sizeClasses[size];

    return (
      <div ref={ref} className={cn(config.container, className)} {...props}>
        <div className={cn('flex items-center justify-center', config.gap, config.spacing)}>
          {icon && <Icon icon={icon} className={cn(config.icon, 'text-slate-600 dark:text-slate-400')} />}
          <p className={cn('text-slate-500 dark:text-slate-400', config.label)}>
            {label}
          </p>
        </div>
        <p className={cn('font-semibold text-slate-900 dark:text-white', config.value)}>
          {value}
        </p>
      </div>
    );
  }
);

StatDisplay.displayName = 'StatDisplay';

export { StatDisplay };