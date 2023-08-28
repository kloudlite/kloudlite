import { getMetadata } from './common';

export const getConfig = (
  { metadata, data, displayName } = {
    metadata: getMetadata(),
    displayName: '',
    data: {},
  }
) => ({
  ...{
    metadata,
    displayName,
    data,
  },
});
