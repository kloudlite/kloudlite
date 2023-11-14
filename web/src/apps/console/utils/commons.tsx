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

export const popupWindow = ({
  url = '',
  onClose = () => {},
  width = 800,
  height = 500,
  title = 'kloudlite',
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
