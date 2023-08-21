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

export const useHandleFromMatches = (key, def = null) => {
  const matches = useMatches();
  const res = useCallback(() => {
    return matches
      .slice()
      .reverse()
      .find((match) => match.handle?.[key]);
  }, [matches])();
  if (res) {
    // console.log(res.handle[key]);
    return res.handle[key];
  }
  return def;
};

export const useDataFromMatches = (key, def = null) => {
  const matches = useMatches();
  const res = matches.reverse().find((match) => match.data?.[key]);
  if (res) {
    return res.data[key];
  }
  return def;
};

// export const useDataFromMatchesWithId = (id) => {
//   const matches = useMatches();
//   const res = matches.find((m) => m.id === id);
//   if (res) {
//     return res.data;
//   }
//   return {};
// };
//
export default useMatches;
