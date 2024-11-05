import { ReactNode } from 'react';
import Tooltip from '@kloudlite/design-system/atoms/tooltip';
import TooltipV2 from '@kloudlite/design-system/atoms/tooltipV2';
import { cn } from '@kloudlite/design-system/utils';
import { Truncate } from '~/root/lib/utils/common';

interface IBase {
  className?: string;
  action?: ReactNode;
}

const BaseStyle = 'flex flex-row items-center gap-xl truncate';

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
  noTooltip = false,
}: {
  data?: ReactNode;
  subtitle?: ReactNode;
  noTooltip?: boolean;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-col flex-1 truncate">
        {noTooltip ? (
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
        ) : (
          <Tooltip.Root
            className="!w-fit !max-w-[500px]"
            side="top"
            content={
              <div className="flex flex-col flex-1">
                {data && (
                  <div className="flex-1 bodyMd-medium text-text-strong pulsable whitespace-normal">
                    {data}
                  </div>
                )}
                {subtitle && (
                  <div className="pulsable bodyMd text-text-soft">
                    {subtitle}
                  </div>
                )}
              </div>
            }
          >
            <div className="flex flex-col gap-sm truncate max-w-full w-fit">
              {data && (
                <div className="flex-1 bodyMd-medium text-text-strong truncate pulsable group is-data">
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
        )}
      </div>
      {action}
    </div>
  );
};

const ListItemV2 = ({
  data,
  subtitle,
  className = '',
  action,
  noTooltip = false,
  truncateLength = 20,
  pre,
  subtitleClass,
  titleClass,
}: {
  pre?: ReactNode;
  data?: string;
  subtitle?: string;
  noTooltip?: boolean;
  truncateLength?: number;
  titleClass?: string;
  subtitleClass?: string;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      {pre && <div className="flex-shrink-0">{pre}</div>}
      <div className="flex flex-col flex-1">
        {noTooltip ? (
          <div className="flex flex-col gap-sm max-w-full w-fit pulsable">
            {data && (
              <div
                className={cn(
                  'flex-1 bodyMd-medium text-text-strong pulsable',
                  titleClass
                )}
              >
                <Truncate length={truncateLength}>{data}</Truncate>
              </div>
            )}
            {subtitle && (
              <div
                className={cn('pulsable bodyMd text-text-soft', subtitleClass)}
              >
                <Truncate length={truncateLength}>{subtitle}</Truncate>
              </div>
            )}
          </div>
        ) : (
          <div className="flex flex-col flex-1">
            {data && (
              <div
                className={cn(
                  'flex-1 bodyMd-medium text-text-strong pulsable whitespace-normal',
                  titleClass
                )}
              >
                {data.length >= truncateLength ? (
                  <TooltipV2 content={data}>
                    <Truncate length={truncateLength}>{data}</Truncate>
                  </TooltipV2>
                ) : (
                  <Truncate length={truncateLength}>{data}</Truncate>
                )}
              </div>
            )}
            {subtitle && (
              <div
                className={cn('pulsable bodyMd text-text-soft', subtitleClass)}
              >
                {subtitle.length >= truncateLength ? (
                  <TooltipV2 content={subtitle}>
                    <Truncate length={truncateLength}>{subtitle}</Truncate>
                  </TooltipV2>
                ) : (
                  <Truncate length={truncateLength}>{subtitle}</Truncate>
                )}
              </div>
            )}
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

const ListTitleV2 = ({
  className,
  action,
  title,
  avatar,
  subtitle,
  truncateLength = 20,
  titleClass,
  subtitleClass,
}: {
  className?: string;
  action?: ReactNode;
  title?: string;
  subtitle?: string;
  avatar?: ReactNode;
  truncateLength?: number;
  titleClass?: string;
  subtitleClass?: string;
}) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-row items-center gap-xl flex-1 truncate">
        {avatar}
        <div className="flex flex-col gap-sm flex-1">
          {title && (
            <div className="bodyMd-semibold text-text-default pulsable">
              <span className={cn('w-fit', titleClass)}>
                {title.length >= truncateLength ? (
                  <TooltipV2 content={title}>
                    <Truncate length={truncateLength}>{title}</Truncate>
                  </TooltipV2>
                ) : (
                  <Truncate length={truncateLength}>{title}</Truncate>
                )}
              </span>
            </div>
          )}

          {subtitle && (
            <div
              className={cn('bodySm text-text-soft pulsable', subtitleClass)}
            >
              {subtitle.length >= truncateLength ? (
                <TooltipV2 content={subtitle}>
                  <Truncate length={truncateLength}>{subtitle}</Truncate>
                </TooltipV2>
              ) : (
                <Truncate length={truncateLength}>{subtitle}</Truncate>
              )}
            </div>
          )}
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
  title: 'w-[80px] flex flex-1',
  author: 'w-[180px]',
  updated: 'w-[180px]',
  status: 'flex-1 min-w-[30px] w-fit',
  action: 'w-[24px]',
  flex: 'flex-1',
  item: 'w-[146px]',
};
export {
  ListBody, listClass, listFlex, ListItem, ListItemV2,
  ListSecondary, ListTitle,
  ListTitleV2
};

