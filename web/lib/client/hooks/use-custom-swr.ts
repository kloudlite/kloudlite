import useSWR from 'swr';
import { IGqlReturn } from '../../types/common';
import { handleError } from '../../utils/common';

const caller = <T>(
  func: () => IGqlReturn<T>,
  toastError?: boolean
): (() => Promise<T>) => {
  return async () => {
    try {
      const { errors, data } = await func();
      if (errors) {
        throw errors;
      }

      return data;
    } catch (e) {
      if (toastError) {
        handleError(e);
      }
      throw e;
    }
  };
};
const useCustomSwr = <T>(
  key: string | null | (() => string | null),
  func: () => IGqlReturn<T>,
  toastError?: boolean
) => {
  return useSWR(key, caller(func, toastError), {
    shouldRetryOnError: false,
    revalidateOnFocus: false,
  });
};

export default useCustomSwr;
