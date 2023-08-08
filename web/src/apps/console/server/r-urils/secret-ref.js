import { getMetadata } from './common';

export const getSecretRef = (
  { metadata = getMetadata(), stringData = {}, cloudProviderName } = {
    stringData: {},
  }
) => ({
  ...{
    metadata,
    stringData,
    cloudProviderName,
  },
});
