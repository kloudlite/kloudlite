import { useEffect, useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { parseError } from '../../utils/common';
import { IExecutorResp } from '../../server/helpers/execute-query-with-context';

interface IReturn<T> {
  data: T | undefined;
  error: string | undefined;
  isLoading: boolean;
}

export const useApiCall = <A, B>(
  fn: IExecutorResp<A, B>,
  vars: B,
  dep: Array<any> = []
): IReturn<A> => {
  const [_data, setData] = useState<A>();
  const [error, setError] = useState<string>();
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    (async () => {
      console.log('called');
      setIsLoading(true);
      try {
        const { data: __data, errors } = await fn(vars);
        if (errors) {
          throw errors[0];
        }
        setData(__data);
      } catch (err) {
        setError(parseError(err).message);
        console.log(err);
        toast.error(parseError(err).message);
      } finally {
        setIsLoading(false);
      }
    })();
  }, dep);
  return { data: _data, error, isLoading };
};
