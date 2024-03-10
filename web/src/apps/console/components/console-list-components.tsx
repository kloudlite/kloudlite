import { ReactNode, forwardRef } from 'react';
import Tooltip from '~/components/atoms/tooltip';
import { cn } from '~/components/utils';

interface IBase {
  className?: string;
  action?: ReactNode;
}

const BaseStyle = 'flex flex-row items-center gap-xl';

const ListSecondary = ({
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
      <div className="flex flex-row items-center gap-xl flex-1 truncate">
        {avatar}
        <div className="flex flex-col gap-sm flex-1 truncate">
          {title && (
            <div className="bodySm truncate text-text-soft pulsable">
              {title}
            </div>
          )}

          {subtitle && (
            <div className="bodyMd-medium truncate pulsable">{subtitle}</div>
          )}
        </div>
      </div>
      {action}
    </div>
  );
};

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
        <Tooltip.Root
          className="!w-fit !max-w-fit"
          side="top"
          content={
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
          }
        >
          <div className="flex flex-col gap-sm truncate max-w-full w-fit">
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
        </Tooltip.Root>
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
      <div className="flex flex-row items-center gap-xl flex-1 truncate">
        {avatar}
        <div className="flex flex-col gap-sm flex-1 truncate">
          <Tooltip.Root
            className="!w-fit !max-w-fit"
            side="top"
            content={
              <div className="flex flex-col gap-sm w-fit truncate">
                {title && (
                  <div className="bodyMd-semibold text-text-default truncate pulsable">
                    <span className="w-fit">{title}</span>
                  </div>
                )}

                {subtitle && (
                  <div className="bodySm text-text-soft truncate pulsable">
                    {subtitle}
                  </div>
                )}
              </div>
            }
          >
            <div className="flex flex-col gap-sm truncate max-w-full w-fit">
              {title && (
                <div className="bodyMd-semibold text-text-default pulsable truncate">
                  <span>{title}</span>
                </div>
              )}

              {subtitle && (
                <div className="bodySm text-text-soft truncate pulsable">
                  {subtitle}
                </div>
              )}
            </div>
          </Tooltip.Root>
        </div>
      </div>
      {action}
    </div>
  );
};

const listFlex = ({ key }: { key: string }) => ({
  key,
  className: 'flex-1',
  render: () => <div />,
});

const listClass = {
  title: 'w-[180px] min-w-[180px] max-w-[180px] mr-2xl',
  author: 'w-[180px] min-w-[180px] max-w-[180px]',
};
export { ListBody, ListItem, ListTitle, ListSecondary, listFlex, listClass };
