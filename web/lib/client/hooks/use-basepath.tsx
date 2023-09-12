import { useLocation } from '@remix-run/react';
import { useCallback } from 'react';

const useBasepath = () => {
  const mainLocation = useLocation();

  const getPath = useCallback(() => {
    const tempPath = mainLocation.pathname.substring(
      0,
      mainLocation.pathname.lastIndexOf('/')
    );
    return mainLocation.pathname.endsWith('/')
      ? tempPath.substring(0, tempPath.lastIndexOf('/'))
      : tempPath;
  }, [mainLocation]);

  const location = () => {
    const getPrevious = () => {
      const prevPath = getPath();
      return { path: prevPath, getPrevious };
    };
    return {
      path: getPath(),
      getPrevious,
    };
  };

  return location();
};

export default useBasepath;
