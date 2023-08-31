import { useMatches as useAbc } from '@remix-run/react';
import { useCallback } from 'react';

const useMatches = () => {
  const matches = useAbc();
  matches.forEach((match) => {
    if (typeof match?.handle === 'function') {
      match.handle = match.handle(match.data);
    }
  });
  return matches;
};

export const useHandleFromMatches = (key: string, def: any = null) => {
  const matches = useMatches();
  const res = useCallback(() => {
    return matches
      .slice()
      .reverse()
      .find((match) => match.handle?.[key]);
  }, [matches])();
  if (res) {
    return res?.handle?.[key];
  }
  return def;
};

export const useDataFromMatches = (key: string, def: any = null) => {
  const matches = useMatches();
  const res = matches
    .slice()
    .reverse()
    .find((match) => match.data?.[key]);
  if (res) {
    return res.data[key];
  }
  return def;
};

export default useMatches;
