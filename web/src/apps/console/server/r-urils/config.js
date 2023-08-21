import { getMetadata } from './common';

export const getConfig = (
  { metadata, data } = {
    metadata: getMetadata(),
    data: {},
  }
) => ({
  ...{
    metadata,
    data,
  },
});
