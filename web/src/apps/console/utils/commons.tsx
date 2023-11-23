import { Badge } from '~/components/atoms/badge';
import { AWSlogoFill, Info } from '@jengaicons/react';
import {
  Kloudlite__Io___Pkg___Types__SyncStatusState as SyncState,
  Kloudlite__Io___Pkg___Types__SyncStatusAction as SyncAction,
  Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider as CloudProviders,
} from '~/root/src/generated/gql/server';
import {
  IManagedServiceTemplate,
  IManagedServiceTemplates,
} from '../server/gql/queries/managed-service-queries';

export const getManagedTemplate = ({
  templates,
  kind,
  apiVersion,
}: {
  templates: IManagedServiceTemplates;
  kind: string;
  apiVersion: string;
}): IManagedServiceTemplate | undefined => {
  return templates
    ?.flatMap((t) => t.items.flat())
    .find((t) => t.kind === kind && t.apiVersion === apiVersion);
};

export const DIALOG_TYPE = Object.freeze({
  ADD: 'add',
  EDIT: 'edit',
  NONE: 'none',
});

export const DIALOG_DATA_NONE = Object.freeze({
  type: DIALOG_TYPE.NONE,
  data: null,
});

export const ACCOUNT_ROLES = Object.freeze({
  account_member: 'Member',
  account_admin: 'Admin',
});

interface IPopupWindowOptions {
  url: string;
  width?: number;
  height?: number;
  title?: string;
}

export const popupWindow = ({
  url,
  onClose = () => {},
  width = 800,
  height = 500,
  title = 'kloudlite',
}: IPopupWindowOptions & {
  onClose?: () => void;
}) => {
  const frame = window.open(
    url,
    title,
    `toolbar=no,scrollbars=yes,resizable=no,top=${
      window.screen.height / 2 - height / 2
    },left=${window.screen.width / 2 - width / 2},width=800,height=600`
  );

  const interval = setInterval(() => {
    if (frame && frame.closed) {
      clearInterval(interval);
      onClose();
    }
  }, 100);
};

export const asyncPopupWindow = (options: IPopupWindowOptions) => {
  return new Promise((resolve) => {
    popupWindow({
      ...options,
      onClose: () => {
        resolve(true);
      },
    });
  });
};

// Component for Status parsing
type IStatus = 'deleting' | 'none';

interface IStatusFormat {
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

export const parseStatus = (item: IStatusFormat) => {
  let status: IStatus = 'none';

  if (item.markedForDeletion) {
    status = 'deleting';
  }

  return { status, component: parseStatusComponent({ status }) };
};

export const downloadFile = ({
  filename,
  data,
  format,
}: {
  filename: string;
  format: string;
  data: string;
}) => {
  const blob = new Blob([data], { type: format });

  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = filename;

  document.body.appendChild(link);

  link.click();

  URL.revokeObjectURL(url);
  document.body.removeChild(link);
};

export const providerIcons = (iconsSize = 16) => {
  return { aws: <AWSlogoFill size={iconsSize} /> };
};

export const renderCloudProvider = ({
  cloudprovider,
}: {
  cloudprovider: CloudProviders | 'unknown';
}) => {
  const iconSize = 16;
  switch (cloudprovider) {
    case 'aws':
      return (
        <div className="flex flex-row gap-xl items-center">
          {providerIcons(iconSize).aws}
          <span>{cloudprovider}</span>
        </div>
      );
    default:
      return cloudprovider;
  }
};
