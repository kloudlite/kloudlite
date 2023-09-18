import { ReactNode } from 'react';
import { Button, IButton } from '~/components/atoms/button';
import { cn } from '~/components/utils';

interface EmptyStateProps {
  image: ReactNode;
  heading: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
  action?: IButton;
  secondaryAction?: IButton;
  shadow?: boolean;
  border?: boolean;
  compact?: boolean;
}

export const EmptyState = ({
  image = null,
  heading = 'This is where you’ll manage your projects',
  children = null,
  footer = null,
  action,
  secondaryAction,
  shadow = true,
  border = true,
  compact = false,
}: EmptyStateProps) => {
  return (
    <div
      className={cn(
        'flex flex-col items-center px-3xl py-8xl rounded bg-surface-basic-default',
        {
          'shadow-button': shadow,
          'border border-border-disabled': border,
          'gap-2xl': compact,
          'gap-5xl': !compact,
        }
      )}
    >
      {image && image}
      <div className="flex flex-col gap-2xl pb-8xl">
        {heading && <div className="headingLg text-center">{heading}</div>}
        {children && (
          <div className="text-text-strong bodyMd text-center">{children}</div>
        )}
        {(action || secondaryAction) && (
          <div className="flex flex-row items-center justify-center gap-lg pt-lg">
            {secondaryAction && <Button {...secondaryAction} />}
            {action && <Button {...action} />}
          </div>
        )}
        {footer && <div className="bodySm text-text-soft">{footer}</div>}
      </div>
    </div>
  );
};
