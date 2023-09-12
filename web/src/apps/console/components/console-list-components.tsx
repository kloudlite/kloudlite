import { ReactNode } from 'react';
import { cn } from '~/components/utils';

const ListBody = ({
  data,
  className = '',
}: {
  data: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('bodyMd text-text-strong truncate', className)}>
      {data}
    </div>
  );
};
const ListItem = ({
  data,
  className = '',
}: {
  data: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('bodyMd-medium text-text-strong', className)}>
      {data}
    </div>
  );
};

const ListItemWithSubtitle = ({
  data,
  subtitle,
  className = '',
}: {
  data: ReactNode;
  subtitle: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('flex flex-col', className)}>
      <ListItem data={data} />
      <div className="bodyMd text-text-soft">{subtitle}</div>
    </div>
  );
};

const ListTitle = ({
  title,
  className = '',
}: {
  title: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('bodyMd-semibold text-text-default', className)}>
      {title}
    </div>
  );
};

const ListTitleWithSubtitle = ({
  title,
  subtitle,
  className = '',
}: {
  title: ReactNode;
  subtitle: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('flex flex-col gap-sm', className)}>
      <ListTitle title={title} />
      <div className="bodySm text-text-soft">{subtitle}</div>
    </div>
  );
};

const ListTitleWithAvatar = ({
  title,
  avatar: Avatar,
  className = '',
}: {
  title: ReactNode;
  avatar: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('flex flex-row items-center gap-lg', className)}>
      {Avatar}
      <ListTitle title={title} />
    </div>
  );
};

const ListTitleWithSubtitleAvatar = ({
  title,
  subtitle,
  avatar: Avatar,
  className = '',
}: {
  title: ReactNode;
  subtitle: ReactNode;
  avatar: ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn('flex flex-row items-center gap-xl', className)}>
      {Avatar}
      <ListTitleWithSubtitle title={title} subtitle={subtitle} />
    </div>
  );
};

export {
  ListBody,
  ListItem,
  ListItemWithSubtitle,
  ListTitle,
  ListTitleWithSubtitle,
  ListTitleWithAvatar,
  ListTitleWithSubtitleAvatar,
};
