import { useLocation } from '@remix-run/react';

export const useActivePath = (conf = {}) => {
  const { parent = '' } = conf;
  const history = useLocation();
  const pathname = history.pathname.toLowerCase() || '';
  if (!parent) {
    return {
      activePath: pathname,
      isBase: pathname === '',
    };
  }
  const parentLowerCase = parent.toLowerCase();
  const splits = pathname.split(parentLowerCase);
  if (splits.length < 1) {
    return undefined;
  }
  const match = splits[1]?.endsWith('/') ? splits[1].slice(0, -1) : splits[1];
  return {
    activePath: match,
    isBase: match === '',
  };
};
