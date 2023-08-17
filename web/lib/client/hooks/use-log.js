import { useEffect } from 'react';

export const useLog = (data) => {
  useEffect(() => {
    console.log(data);
  }, [data]);
};
