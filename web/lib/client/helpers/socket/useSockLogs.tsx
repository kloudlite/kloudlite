import { useEffect, useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';
import { ISocketResp, useSubscribe } from './context';

interface IuseLog {
  account: string;
  cluster: string;
  trackingId: string;
}

export const useSocketLogs = ({ account, cluster, trackingId }: IuseLog) => {
  const [logs, setLogs] = useState<ISocketResp<IuseLog>[]>([]);
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

  useEffect(() => {
    const sorted = responses.sort((a, b) => {
      const resp = b.data.podName.localeCompare(a.data.podName);

      if (resp === 0) {
        return dayjs(a.data.timestamp).unix() - dayjs(b.data.timestamp).unix();
      }

      return resp;
    });

    if (JSON.stringify(sorted) !== JSON.stringify(logs)) {
      setLogs(sorted);
    }
  }, [responses]);

  return {
    logs,
    errors,
    isLoading,
    subscribed,
  };
};
