import { useEffect } from 'react';
import { ISocketResp, useSubscribe } from './context';
import { useReload } from '../reloader';

export const useSocketWatch = (
  onUpdate: (v: ISocketResp<any>[]) => void,
  topic: string
) => {
  const { responses, subscribed } = useSubscribe(
    {
      for: 'resource-update',
      data: {
        id: topic,
        respath: topic,
      },
    },
    []
  );

  useEffect(() => {
    if (subscribed) {
      onUpdate(responses);
    }
  }, [responses]);
};

export const useWatchReload = (topic: string) => {
  const reloadPage = useReload();
  useSocketWatch((rd) => {
    console.log(rd);
    if (rd.find((v) => v.id === topic)) {
      console.log('reloading due to watch event', rd);
      reloadPage();
    }
  }, topic);
};
