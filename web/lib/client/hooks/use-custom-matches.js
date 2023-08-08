import { useMatches as useAbc } from '@remix-run/react';

const useMatches = () => {
  const matches = useAbc();
  matches.forEach((match) => {
    if (typeof match?.handle === 'function') {
      match.handle = match.handle(match.data);
    }
  });
  return matches;
};

export const useLoaderDataFromMatches = (key) => {
  const matches = useMatches();
  return matches.reverse().find((match) => match.handle[key]);
};

export default useMatches;
