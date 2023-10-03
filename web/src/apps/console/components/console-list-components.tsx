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
      <div className="flex-1 truncate">{data}</div>
      {action}
    </div>
  );
};
const ListItem = ({
  data,
  className = '',
  action,
}: {
  data: ReactNode;
} & IBase) => {
  return (
    <div
      className={cn(
        'bodyMd-medium text-text-strong truncate',
        BaseStyle,
        className
      )}
    >
      <div className="flex-1 truncate">{data}</div>
      {action}
    </div>
  );
};

const ListItemWithSubtitle = ({
  data,
  subtitle,
  className = '',
  action,
}: {
  data: ReactNode;
  subtitle: ReactNode;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-col flex-1 truncate">
        <ListItem data={data} />
        {subtitle && (
          <div className="bodyMd text-text-soft truncate">{subtitle}</div>
        )}
      </div>
      {action}
    </div>
  );
};

const ListTitle = ({
  title,
  className = '',
  action,
}: {
  title: ReactNode;
} & IBase) => {
  return (
    <div
      className={cn(
        'bodyMd-semibold text-text-default truncate',
        BaseStyle,
        className
      )}
    >
      <div className="flex-1 truncate">{title}</div>
      {action}
    </div>
  );
};

const ListTitleWithSubtitle = ({
  title,
  subtitle,
  className = '',
  action,
}: {
  title: ReactNode;
  subtitle: ReactNode;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-col gap-sm flex-1 truncate">
        <ListTitle title={title} />
        {subtitle && (
          <div className="bodySm text-text-soft truncate">{subtitle}</div>
        )}
      </div>
      {action}
    </div>
  );
};

const ListTitleWithAvatar = ({
  title,
  avatar: Avatar,
  className = '',
  action,
}: {
  title: ReactNode;
  avatar: ReactNode;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-row items-center gap-lg flex-1">
        {Avatar}
        <ListTitle title={title} />
      </div>
      {action}
    </div>
  );
};

const ListTitleWithSubtitleAvatar = ({
  title,
  subtitle,
  avatar: Avatar,
  className = '',
  action,
}: {
  title: ReactNode;
  subtitle: ReactNode;
  avatar: ReactNode;
} & IBase) => {
  return (
    <div className={cn(BaseStyle, className)}>
      <div className="flex flex-row items-center gap-xl flex-1">
        {Avatar}
        <ListTitleWithSubtitle title={title} subtitle={subtitle} />
      </div>
      {action}
    </div>
  );
};

export {
  ListBody,
  ListItem,
  ListItemWithSubtitle,
  ListTitle,
  ListTitleWithAvatar,
  ListTitleWithSubtitle,
  ListTitleWithSubtitleAvatar
};

