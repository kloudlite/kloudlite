import { useEffect } from 'react';

export const useLog = (data: any) => {
  useEffect(() => {
    console.log(data);
  }, [data]);
};
