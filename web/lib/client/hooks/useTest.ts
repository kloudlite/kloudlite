import { useEffect } from 'react';

export const useTest = (data: number) => {
  useEffect(() => {
    console.log(data);
  }, [data]);
};
