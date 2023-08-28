import { getMetadata } from './common';

export const getProjectSepc = (
  { targetNamespace, clusterName, accountName } = {
    targetNamespace: '',
    clusterName: '',
    accountName: '',
  }
) => ({
  ...{
    targetNamespace,

    clusterName,
    accountName,
  },
});

export const getProject = (
  { metadata, spec, displayName } = {
    metadata: getMetadata(),
    spec: getProjectSepc(),
    displayName: '',
  }
) => ({
  ...{
    metadata,
    displayName,
    spec,
  },
});
