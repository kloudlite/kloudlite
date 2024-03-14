import { useEffect, useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';
import { ISocketResp, useSubscribe } from './context';
import { ILog } from '../../components/logger';

interface IuseLog {
  account: string;
  cluster: string;
  trackingId: string;
  recordVersion?: number;
}

export const useSocketLogs = ({
  account,
  cluster,
  trackingId,
  recordVersion,
}: IuseLog) => {
  const [logs, setLogs] = useState<ISocketResp<ILog>[]>([]);
  const { responses, infos, subscribed, errors } = useSubscribe(
    {
      for: 'logs',
      data: {
        id: `${account}.${cluster}.${trackingId}`,
        spec: {
          ...{
            account,
            cluster,
            trackingId,
            recordVersion,
          },
        },
      },
    },
    []
  );

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
    isLoading: !subscribed && (logs.length === 0 || infos.length === 0),
    subscribed,
  };
};
