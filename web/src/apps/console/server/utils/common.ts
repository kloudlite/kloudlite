import { decodeUrl } from '~/root/lib/client/hooks/use-search';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IRemixCtx } from '~/root/lib/types/common';

export interface IHandleProps<T = boolean> {
  show: T;
  setShow: (fn: T) => void;
}

export const getPagination = (ctx: IRemixCtx) => {
  const { page } = getQueries(ctx);
  const { orderBy, sortDirection, last, first, before, after } =
    decodeUrl(page);

  return {
    ...{
      orderBy: orderBy || 'updateTime',
      sortDirection: sortDirection || 'DESC',
      last,
      first: first || (last ? undefined : 10),
      before,
      after,
    },
  };
};

export const getSearch = (ctx: IRemixCtx) => {
  const { search } = getQueries(ctx);
  const s = decodeUrl(search) || {};
  return {
    ...s,
  };
};

export const isValidRegex = (regexString = '') => {
  let isValid = true;
  try {
    RegExp(regexString);
  } catch (e) {
    isValid = false;
  }
  return isValid;
};
