import { ReactNode } from 'react';
import { cn } from '~/components/utils';

interface IBase {
  className?: string;
  action?: ReactNode;
}

const BaseStyle = 'flex flex-row items-center gap-xl';

const ListBody = ({
  data,
  className = '',
  action,
}: {
  data: ReactNode;
} & IBase) => {
  return (
    <div
      className={cn('bodyMd text-text-strong truncate', BaseStyle, className)}
    >
      <div className="flex-1 truncate pulsable">{data}</div>
      {action}
    </div>
  );
};

const ListItem = ({
  data,
  subtitle,
  className = '',
  action,
}: {
  data?: ReactNode;
  subtitle?: ReactNode;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-col flex-1 truncate">
        {data && (
          <div className="flex-1 bodyMd-medium text-text-strong truncate pulsable">
            {data}
          </div>
        )}
        {subtitle && (
          <div className="pulsable bodyMd text-text-soft truncate">
            {subtitle}
          </div>
        )}
      </div>
      {action}
    </div>
  );
};

const ListTitle = ({
  className,
  action,
  title,
  avatar,
  subtitle,
}: {
  className?: string;
  action?: ReactNode;
  title?: ReactNode;
  subtitle?: ReactNode;
  avatar?: ReactNode;
}) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-row items-center gap-xl flex-1">
        {avatar}
        <div className="flex flex-col gap-sm flex-1 truncate">
          {title && (
            <div className="bodyMd-semibold text-text-default truncate pulsable">
              {title}
            </div>
          )}

          {subtitle && (
            <div className="bodySm text-text-soft truncate pulsable">
              {subtitle}
            </div>
          )}
        </div>
      </div>
      {action}
    </div>
  );
};

export { ListBody, ListItem, ListTitle };
