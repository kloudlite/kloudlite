import { getMetadata } from './common';

export const getSecretRef = (
  { metadata = getMetadata(), stringData = {} } = {
    stringData: {},
  }
) => ({
  ...{
    metadata,
    stringData,
  },
});
