import { getMetadata } from './common';

export const getClusterSepc = ({
  accountName,
  config,
  count = 1,
  provider,
  providerName,
  region,
} = {}) => ({
  ...{
    accountName,
    config,
    count,
    provider,
    providerName,
    region,
  },
});

export const getCluster = ({
  metadata = getMetadata(),
  spec = getClusterSepc(),
} = {}) => ({
  ...{
    metadata,
    spec,
  },
});
