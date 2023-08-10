import { getMetadata } from './common';

export const getCredentialsRef = ({ name, namespace }) => ({
  ...{ name, namespace },
});

export const getClusterSepc = (
  {
    accountName,
    region,
    cloudProvider,
    credentialsRef,
    vpc,
    availabilityMode,
  } = {
    accountName: '',
    region: '',
    cloudProvider: 'aws',
    credentialsRef: getCredentialsRef(),
    vpc: '',
    availabilityMode: 'HA' || 'dev',
  }
) => ({
  ...{
    cloudProvider,
    credentialsRef,
    vpc,
    region,
    availabilityMode,
    accountName,
  },
});

export const getCluster = (
  { metadata, spec } = {
    metadata: getMetadata(),
    spec: getClusterSepc(),
  }
) => ({
  ...{
    metadata,
    spec,
  },
});
