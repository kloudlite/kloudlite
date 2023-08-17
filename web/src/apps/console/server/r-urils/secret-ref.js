import { getMetadata } from './common';

export const getSecretRef = (
  { metadata, stringData, cloudProviderName } = {
    metadata: getMetadata(),
    stringData: {},
    cloudProviderName: '',
  }
) => ({
  ...{
    metadata,
    stringData,
    cloudProviderName,
  },
});
