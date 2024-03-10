import { useCallback } from 'react';
import { ISocketResp, useSubscribe } from './context';
import { useReload } from '../reloader';
import useDebounce from '../../hooks/use-debounce';

export const useSocketWatch = (
  onUpdate: (v: ISocketResp<any>[]) => void,
  topic: string | string[]
) => {
  const { responses, subscribed } = useSubscribe(
    Array.isArray(topic)
      ? topic.map((t) => {
          return {
            for: 'resource-update',
            data: {
              id: t,
              respath: t,
            },
          };
        })
      : {
          for: 'resource-update',
          data: {
            id: topic,
            respath: topic,
          },
        },
    []
  );

  useDebounce(
    () => {
      if (subscribed) {
        onUpdate(responses);
      }
    },
    1000,
    [responses]
  );
};

export const useWatchReload = (topic: string | string[]) => {
  const reloadPage = useReload();
  const topicMap: {
    [key: string]: boolean;
  } = useCallback(
    () =>
      Array.isArray(topic)
        ? topic.reduce((acc, curr) => {
            return { ...acc, [curr]: true };
          }, {})
        : { [topic]: true },
    [topic]
  )();

  useSocketWatch((rd) => {
    if (rd.find((v) => topicMap[v.id])) {
      console.log('reloading due to watch event');
      reloadPage();
    }
  }, topic);
};
