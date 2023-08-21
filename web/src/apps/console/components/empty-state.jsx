import PropTypes from 'prop-types';
import { Button } from '~/components/atoms/button';
import { cn } from '~/components/utils';

export const EmptyState = ({
  image = null,
  heading,
  children = null,
  footer = null,
  action = null,
  secondaryAction = null,
}) => {
  return (
    <div
      className={cn(
        'flex flex-col items-center px-3xl py-8xl gap-5xl shadow-button border border-border-disabled rounded bg-surface-basic-default'
      )}
    >
      {image && image}
      <div className="flex flex-col gap-2xl pb-8xl">
        <div className="headingLg text-center">{heading}</div>
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

EmptyState.propTypes = {
  heading: PropTypes.string,
};

EmptyState.defaultProps = {
  heading: 'This is where you’ll manage your projects',
};
