import { kindv } from './api-versions';
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
    apiVersion: kindv.cluster.apiVersion,
    kind: kindv.cluster.kind,
    metadata,
    spec,
  },
});
