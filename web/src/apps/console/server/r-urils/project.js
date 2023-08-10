import { getMetadata } from './common';

export const getProjectSepc = (
  { targetNamespace, displayName, clusterName, accountName } = {
    targetNamespace: '',
    displayName: '',
    clusterName: '',
    accountName: '',
  }
) => ({
  ...{
    targetNamespace,
    displayName,
    clusterName,
    accountName,
  },
});

export const getProject = (
  { metadata, spec } = {
    metadata: getMetadata(),
    spec: getProjectSepc(),
  }
) => ({
  ...{
    metadata,
    spec,
  },
});
