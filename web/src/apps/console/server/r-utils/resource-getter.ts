import { AppIn } from '~/root/src/generated/gql/server';

export const getAppIn = ({
  spec,
  metadata,
  displayName,
  kind,
  apiVersion,
  enabled,
}: AppIn): AppIn => {
  return {
    spec,
    metadata,
    displayName,
    kind,
    apiVersion,
    enabled,
  };
};
