import { useEffect } from 'react';

export const useLog = (data: any) => {
  useEffect(() => {
    console.trace(data);
  }, [data]);
};
