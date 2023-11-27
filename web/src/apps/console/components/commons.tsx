import { CopySimple, Check, Info } from '@jengaicons/react';
import { ReactNode, useState } from 'react';
import Chips from '~/components/atoms/chips';
import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { generateKey, titleCase } from '~/components/utils';
import { Badge } from '~/components/atoms/badge';
import {
  Kloudlite__Io___Pkg___Types__SyncStatusState as SyncState,
  Kloudlite__Io___Pkg___Types__SyncStatusAction as SyncAction,
} from '~/root/src/generated/gql/server';
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
      <div className="bodyMd text-text-strong">{value}</div>
    </div>
  );
};

export const CopyButton = ({
  title,
  value,
}: {
  title: ReactNode;
  value: string;
}) => {
  const [copyIcon, setCopyIcon] = useState(<CopySimple />);
  const { copy } = useClipboard({
    onSuccess: () => {
      setTimeout(() => {
        setCopyIcon(<CopySimple />);
      }, 1000);
      toast.success('Copied to clipboard');
    },
  });

  return (
    <Chips.Chip
      type="CLICKABLE"
      item={title}
      label={title}
      prefix={copyIcon}
      onClick={() => {
        copy(value);
        setCopyIcon(<Check />);
      }}
    />
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
type IStatus = 'deleting' | 'none';

interface IStatusMeta {
  markedForDeletion?: boolean;
  status?: {
    checks?: any;
    isReady: boolean;
    message?: { RawMessage?: any };
  };
  syncStatus: {
    action: SyncAction;
    error?: string;
    state: SyncState;
  };
}

type ICommonMeta = IUpdateMeta & IStatusMeta;

const parseStatusComponent = ({ status }: { status: IStatus }) => {
  switch (status) {
    case 'deleting':
      return (
        <Badge icon={<Info />} type="critical">
          Deleting...
        </Badge>
      );
    default:
      return null;
  }
};

export const parseStatus = (item: IStatusMeta) => {
  let status: IStatus = 'none';

  if (item.markedForDeletion) {
    status = 'deleting';
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
    statusRender: ({ className }: { className: string }) => {
      return {
        key: generateKey(keyPrefix, 'status'),
        className,
        render: () => parseStatus(resource).component,
      };
    },
  };
};
