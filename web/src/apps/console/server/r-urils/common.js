import getQueries from '~/root/lib/server/helpers/get-queries';
import { decodeUrl } from '~/root/lib/client/hooks/use-search';
import { keyconstants } from './key-constants';

export const getMetadata = (
  { name, labels = {}, annotations = {}, namespace = undefined } = {
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

export const parseName = (resource = {}) => resource?.metadata?.name || '';
export const parseCreationTime = (resource = {}) =>
  resource?.creationTime || '';
export const parseUpdationTime = (resource = {}) => resource?.updateTime || '';

export const parseDisplaynameFromAnn = (resource = {}) =>
  resource?.metadata?.annotations?.[keyconstants.displayName] || '';

export const parseFromAnn = (resource = {}, key = '') =>
  resource?.metadata?.annotations?.[key] || '';

export const newPagination = ({
  orderBy,
  sortBy,
  last,
  first,
  before,
  after,
}) => {
  return {
    ...{
      orderBy,
      sortBy,
      last,
      first,
      before,
      after,
    },
  };
};

export const getPagination = (ctx = {}) => {
  const { page } = getQueries(ctx);
  const { orderBy, sortDirection, last, first, before, after } =
    decodeUrl(page);
  return {
    ...{
      orderBy,
      sortDirection,
      last,
      first,
      before,
      after,
    },
  };
};

export const getSearch = (ctx = {}) => {
  const { search } = getQueries(ctx);
  const s = decodeUrl(search) || {};
  return {
    ...s,
  };
};

export const parseStatus = (item) => {
  return item?.status;
};
