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
  { metadata, spec } = {
    metadata: getMetadata(),
    spec: getWorkspaceSpecs(),
  }
) => ({
  ...{
    spec,
    metadata,
  },
});
