import { useEffect, useState } from 'react';
import { useSubscribe } from './context';

interface IuseLog {
  account: string;
  cluster: string;
  trackingId: string;
}

export const useSocketLogs = ({ account, cluster, trackingId }: IuseLog) => {
  const { responses, subscribed, errors } = useSubscribe(
    {
      for: 'logs',
      data: {
        id: `${account}.${cluster}.${trackingId}`,
        spec: {
          account,
          cluster,
          trackingId,
        },
      },
    },
    []
  );

  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (subscribed && isLoading) {
      setIsLoading(false);
    } else if (!subscribed && !isLoading) {
      setIsLoading(true);
    }
  }, []);

  return {
    logs: responses,
    errors,
    isLoading,
    subscribed,
  };
};
