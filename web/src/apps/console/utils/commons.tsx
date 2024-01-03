import { AWSlogoFill } from '@jengaicons/react';
import { Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider as CloudProviders } from '~/root/src/generated/gql/server';
import {
  IMSvTemplate,
  IMSvTemplates,
} from '../server/gql/queries/managed-templates-queries';

export const getManagedTemplate = ({
  templates,
  kind,
  apiVersion,
}: {
  templates: IMSvTemplates;
  kind: string;
  apiVersion: string;
}): IMSvTemplate | undefined => {
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

export const flatMap = (data: any) => {
  const keys = data.split('.');

  const jsonObject = keys.reduceRight(
    (acc: any, key: string) => ({ [key]: acc }),
    null
  );
  return jsonObject;
};
