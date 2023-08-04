import { keyconstants } from './key-constants';

export const getMetadata = (
  { name, labels = {}, annotations = {}, namespace } = {
    name: '',
  }
) => ({
  ...{
    name,
    labels,
    annotations,
    namespace,
  },
});

export const parseName = (resource) => resource?.metadata?.name || '';
export const parseCreationTime = (resource) => resource?.creationTime || '';
export const parseUpdationTime = (resource) => resource?.updateTime || '';

export const parseDisplaynameFromAnn = (resource) =>
  resource?.metadata?.annotations?.[keyconstants.displayName] || '';

export const parseFromAnn = (resource, key) =>
  resource?.metadata?.annotations?.[key] || '';

export const getPagination = ({
  orderBy,
  sortBy,
  last,
  first,
  before,
  after,
}) => ({
  ...{
    orderBy,
    sortBy,
    last,
    first,
    before,
    after,
  },
});
