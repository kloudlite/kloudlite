import { useEffect } from 'react';

export const useLog = (data) => {
  useEffect(() => {
    console.trace(data);
  }, [data]);
};
