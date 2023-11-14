import useSWR from 'swr';
import { IGqlReturn } from '../../types/common';

const caller = <T>(func: () => IGqlReturn<T>): (() => Promise<T>) => {
  return async () => {
    const { errors, data } = await func();
    if (errors) {
      throw errors;
    }

    return data;
  };
};
const useCustomSwr = <T>(
  key: string | null | (() => string | null),
  func: () => IGqlReturn<T>
) => {
  return useSWR(key, caller(func));
};

export default useCustomSwr;
