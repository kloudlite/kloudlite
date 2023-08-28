import { getMetadata } from './common';

export const getWorkspaceSpecs = (
  { targetNamespace, projectName } = {
    projectName: '',
    targetNamespace: '',
  }
) => ({
  targetNamespace,
  projectName,
});
export const getWorkspace = (
  { metadata, spec, displayName } = {
    metadata: getMetadata(),
    displayName: '',
    spec: getWorkspaceSpecs(),
  }
) => ({
  ...{
    spec,
    displayName,
    metadata,
  },
});
