import { getMetadata } from './common';

export const getAccount = (
  { metadata, spec = {}, displayName, contactEmail } = {
    displayName: '',
    metadata: getMetadata(),
    contactEmail: '',
  }
) => ({
  ...{
    contactEmail,
    metadata,
    spec,
    displayName,
  },
});
