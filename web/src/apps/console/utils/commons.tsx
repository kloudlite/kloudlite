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
