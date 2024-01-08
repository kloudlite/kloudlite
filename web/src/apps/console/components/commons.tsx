import { CopySimple, Question } from '@jengaicons/react';
import { ReactNode, useState } from 'react';
import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { generateKey, titleCase } from '~/components/utils';
import {
  Github__Com___Kloudlite___Api___Pkg___Types__SyncState as SyncState,
  Github__Com___Kloudlite___Api___Pkg___Types__SyncAction as SyncAction,
} from '~/root/src/generated/gql/server';
import Tooltip from '~/components/atoms/tooltip';
import { Link } from '@remix-run/react';
import { ListItem } from './console-list-components';
import {
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '../server/r-utils/common';

export const BlackProdLogo = ({ size = 16 }) => {
  return <ProdLogo color="currentColor" size={size} />;
};

export const BlackWorkspaceLogo = ({ size = 16 }) => {
  return <WorkspacesLogo color="currentColor" size={size} />;
};

export const DetailItem = ({
  title,
  value,
}: {
  title: ReactNode;
  value: ReactNode;
}) => {
  return (
    <div className="flex flex-col gap-lg flex-1 min-w-[45%]">
      <div className="bodyMd-medium text-text-default">{title}</div>
      <div className="bodyMd text-text-strong w-fit">{value}</div>
    </div>
  );
};

interface InfoLabelProps {
  info: ReactNode;
  label: ReactNode;
  title?: ReactNode;
}

export const InfoLabel = ({ info, title, label }: InfoLabelProps) => {
  return (
    <span className="flex items-center gap-lg">
      {label}{' '}
      <Tooltip.Root
        content={
          <div className="p-md text-xs flex flex-col gap-md">
            <div className="headingSm">{title}</div>
            {info}
          </div>
        }
      >
        <span className="text-text-primary">
          <Question color="currentColor" size={13} />
        </span>
      </Tooltip.Root>
    </span>
  );
};

export const CopyButton = ({
  title,
  value,
}: {
  title: ReactNode;
  value: string;
}) => {
  const [_, setCopyIcon] = useState(<CopySimple />);
  const { copy } = useClipboard({
    onSuccess: () => {
      setTimeout(() => {
        setCopyIcon(<CopySimple />);
      }, 1000);
      // toast.success('Copied to clipboard');
    },
  });

  return (
    // <Chips.Chip
    //   type="CLICKABLE"
    //   item={title}
    //   label={title}
    //   prefix={copyIcon}
    //   onClick={() => {
    //     copy(value);
    //     setCopyIcon(<Check />);
    //   }}
    // />
    <div
      onClick={() => {
        copy(value);
      }}
      className="flex flex-row gap-md items-center select-none group cursor-pointer"
    >
      <span>{title}</span>
      <span className="invisible group-hover:visible">
        <CopySimple size={10} />
      </span>
    </div>
  );
};

interface IUpdateMeta {
  lastUpdatedBy: {
    userName: string;
  };
  createdBy: {
    userName: string;
  };
  updateTime: string;
  creationTime: string;
}

// Component for Status parsing
export type IStatus = 'deleting' | 'notready' | 'syncing' | 'none';

interface IStatusMeta {
  markedForDeletion?: boolean;
  status?: {
    checks?: any;
    isReady: boolean;
    message?: { RawMessage?: any };
  };
  syncStatus?: {
    action?: SyncAction;
    error?: string;
    state?: SyncState;
  };
}

type IResourceType = 'nodepool';

type ICommonMeta = IUpdateMeta & IStatusMeta;

const parseStatusComponent = ({ status }: { status: IStatus }) => {
  switch (status) {
    case 'deleting':
      return <div className="bodyMd text-text-soft pulsable">Deleting...</div>;
    case 'notready':
      return <div className="bodyMd text-text-soft pulsable">Not Ready</div>;
    case 'syncing':
      return <div className="bodyMd text-text-soft pulsable">Syncing</div>;
    default:
      return null;
  }
};

export const parseStatus = ({
  item,
  type,
}: {
  item: IStatusMeta;
  type?: IResourceType;
}) => {
  let status: IStatus = 'none';

  if (item.markedForDeletion) {
    status = 'deleting';
  } else if (!item.status?.isReady) {
    switch (type) {
      case 'nodepool':
        status = 'syncing';
        break;
      default:
        status = 'notready';
    }
  }

  return { status, component: parseStatusComponent({ status }) };
};

export const listRender = ({
  resource,
  keyPrefix,
}: {
  keyPrefix: string;
  resource: ICommonMeta;
}) => {
  return {
    authorRender: ({ className }: { className: string }) => {
      const updateInfo = {
        author: `Updated by ${titleCase(parseUpdateOrCreatedBy(resource))}`,
        time: parseUpdateOrCreatedOn(resource),
      };
      return {
        key: generateKey(keyPrefix, 'author'),
        className,
        render: () => (
          <ListItem data={updateInfo.author} subtitle={updateInfo.time} />
        ),
      };
    },
    statusRender: ({
      className,
      type,
    }: {
      className: string;
      type?: IResourceType;
    }) => {
      return {
        key: generateKey(keyPrefix, 'status'),
        className,
        render: () => parseStatus({ item: resource, type }).component,
        status: parseStatus({ item: resource, type }).status,
      };
    },
  };
};

export const SubHeaderTitle = ({
  to,
  toTitle,
  title,
}: {
  to: string;
  toTitle: ReactNode;
  title: ReactNode;
}) => {
  return (
    <div className="flex flex-col gap-md">
      <Link to={to} className="text-text-soft bodySm">
        {toTitle}
      </Link>
      <span>{title}</span>
    </div>
  );
};
