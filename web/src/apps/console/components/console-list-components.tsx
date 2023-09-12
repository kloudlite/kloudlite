import { ReactNode } from 'react';

const ListBody = ({ data }: { data: ReactNode }) => {
  return <div className="bodyMd text-text-strong truncate">{data}</div>;
};
const ListItem = ({ data }: { data: ReactNode }) => {
  return <div className="bodyMd-medium text-text-strong">{data}</div>;
};

const ListItemWithSubtitle = ({
  data,
  subtitle,
}: {
  data: ReactNode;
  subtitle: ReactNode;
}) => {
  return (
    <div className="flex flex-col">
      <ListItem data={data} />
      <div className="bodyMd text-text-soft">{subtitle}</div>
    </div>
  );
};

const ListTitle = ({ title }: { title: ReactNode }) => {
  return <div className="bodyMd-semibold text-text-default">{title}</div>;
};

const ListTitleWithSubtitle = ({
  title,
  subtitle,
}: {
  title: ReactNode;
  subtitle: ReactNode;
}) => {
  return (
    <div className="flex flex-col gap-sm">
      <ListTitle title={title} />
      <div className="bodySm text-text-soft">{subtitle}</div>
    </div>
  );
};

const ListTitleWithAvatar = ({
  title,
  avatar: Avatar,
}: {
  title: ReactNode;
  avatar: ReactNode;
}) => {
  return (
    <div className="flex flex-row items-center gap-lg">
      {Avatar}
      <ListTitle title={title} />
    </div>
  );
};

const ListTitleWithSubtitleAvatar = ({
  title,
  subtitle,
  avatar: Avatar,
}: {
  title: ReactNode;
  subtitle: ReactNode;
  avatar: ReactNode;
}) => {
  return (
    <div className="flex flex-row items-center gap-xl">
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
