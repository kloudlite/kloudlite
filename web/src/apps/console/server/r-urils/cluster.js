import { getMetadata } from './common';

export const getClusterSepc = ({
  accountName,
  region,
  cloudProvider,
  credentialsRef,
  vpc,
  availabilityMode,
} = {}) => ({
  ...{
    cloudProvider,
    credentialsRef,
    vpc,
    region,
    availabilityMode,
    accountName,
  },
});

export const getCredentialsRef = ({ name, namespace }) => ({
  ...{ name, namespace },
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
