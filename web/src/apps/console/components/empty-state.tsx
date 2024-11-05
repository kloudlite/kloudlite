/* eslint-disable no-nested-ternary */
import { ReactNode, isValidElement } from 'react';
import { Button, IButton } from '@kloudlite/design-system/atoms/button';
import { cn } from '@kloudlite/design-system/utils';

interface EmptyStateProps {
  image: ReactNode;
  heading: ReactNode;
  children?: ReactNode;
  footer?: ReactNode;
  action?: IButton | ReactNode;
  secondaryAction?: IButton;
  shadow?: boolean;
  border?: boolean;
  compact?: boolean;
  padding?: boolean;
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
  padding = true,
}: EmptyStateProps) => {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center px-3xl rounded bg-surface-basic-default',
        {
          'shadow-button': shadow,
          'border border-border-disabled': border,
          'gap-2xl': compact,
          'gap-5xl': !compact,
          'py-8xl': padding,
          'py-lg': !padding,
        }
      )}
    >
      {image && image}
      <div className={cn('flex flex-col gap-2xl', padding ? 'pb-8xl' : '')}>
        {heading && <div className="headingLg text-center">{heading}</div>}
        {children && (
          <div className="text-text-strong bodyMd text-center">{children}</div>
        )}
        {(action || secondaryAction) && (
          <div className="flex flex-row items-center justify-center gap-lg pt-lg">
            {secondaryAction && <Button {...secondaryAction} />}
            {isValidElement(action) ? (
              action
            ) : typeof action === 'object' ? (
              <Button {...(action as IButton)} />
            ) : (
              action
            )}
          </div>
        )}
        {footer && <div className="bodySm text-text-soft">{footer}</div>}
      </div>
    </div>
  );
};
