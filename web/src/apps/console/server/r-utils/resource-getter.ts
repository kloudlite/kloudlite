import { AppIn } from '~/root/src/generated/gql/server';

export const getAppIn = ({
  spec,
  metadata,
  displayName,
  enabled,
}: AppIn): AppIn => {
  return {
    spec,
    metadata,
    displayName,
    enabled,
  };
};
