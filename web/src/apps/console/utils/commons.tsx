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
