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
      const podDiff = a.data.podName.localeCompare(b.data.podName);

      if (podDiff === 0) {
        const contDiff = a.data.containerName.localeCompare(
          b.data.containerName
        );

        if (contDiff === 0) {
          return (
            dayjs(a.data.timestamp).unix() - dayjs(b.data.timestamp).unix()
          );
        }

        return contDiff;
      }

      return podDiff;
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
