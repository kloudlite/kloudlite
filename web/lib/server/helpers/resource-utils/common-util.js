export const getName = (resource) => {
  return resource?.metadata?.name;
};

export const getAnnotations = (annotations = undefined) => {
  if (annotations) return { ...annotations };
  return undefined;
};

export const getLabels = (labels = undefined) => {
  if (labels) return { ...labels };
  return undefined;
};

export const parseLabels = (res) => res?.metadata?.labels || {};
export const parseAnnotations = (res) => res?.metadata?.annotations || {};

export const parseDisplayNameFromAnn = (resource) => {
  return parseAnnotations(resource)['kloudlite.io/display-name'] || '';
};

export const setDisplayNameFromAnn = (resource, name) => {
  return {
    ...(parseAnnotations(resource) || {}),
    'kloudlite.io/display-name': name,
  };
};

export const getMetadata = (
  {
    namespace = null,
    name,
    labels = getLabels(),
    annotations = getAnnotations(),
  } = {
    name: '',
  }
) => ({ ...{ namespace, name, labels, annotations } });

export const copyData = (v) => JSON.parse(JSON.stringify(v));
export const isEqual = (a, b) => JSON.stringify(a) === JSON.stringify(b);
