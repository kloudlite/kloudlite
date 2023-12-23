import { createContext, useContext, useEffect, useMemo } from 'react';
import * as sock from 'websocket';
import { v4 as uuid } from 'uuid';
import { socketUrl } from '~/lib/configs/base-url.cjs';
import { ChildrenProps } from '~/components/types';
import logger from './log';
import { NonNullableString } from '../../types/common';
import { useReload } from './reloader';

type Ievent = 'subscribe' | 'unsubscribe' | NonNullableString;

type IMessage = {
  data: string;
  event: Ievent;
};

const message = ({ event, data }: IMessage): string => {
  return JSON.stringify({ event, data });
};

const socketContextDefaultValue: {
  subscribe: (
    topic: string,
    callback: (arg: { topic: string; message: string }) => void
  ) => string;
  unsubscribe: (topic: string, id: string) => void;
} = {
  subscribe: (): string => '',
  unsubscribe: (): void => {},
};

const createSocketContext = () => {
  if (typeof window === 'undefined') {
    return socketContextDefaultValue;
  }

  const callbacks: {
    [key: string]: {
      [key: string]: (arg: { topic: string; message: string }) => void;
    };
  } = {};

  const wsclient = new Promise<sock.w3cwebsocket>((res, rej) => {
    try {
      // eslint-disable-next-line new-cap
      const w = new sock.w3cwebsocket(socketUrl, '', '', {});

      w.onmessage = (msg) => {
        try {
          const m: {
            topic: string;
            message: string;
            type: 'update' | 'error' | 'info';
          } = JSON.parse(msg.data as string);

          if (m.type === 'error') {
            console.error(m.message);
            return;
          }

          if (m.type === 'info') {
            console.log(m.message);
            return;
          }

          Object.values(callbacks[m.topic]).forEach((cb) => {
            cb(m);
          });
        } catch (err) {
          console.error(err);
        }
      };

      w.onopen = () => {
        res(w);
      };

      w.onerror = (e) => {
        rej(e);
      };

      w.onclose = () => {
        // wsclient.send(newMessage({ event: 'unsubscribe', data: 'test' }));
        logger.log('socket disconnected');
      };
    } catch (e) {
      rej(e);
    }
  });

  return {
    subscribe: (
      topic: string,
      callback: (arg: { topic: string; message: string }) => void
    ): string => {
      (async () => {
        if (!callbacks[topic]) {
          callbacks[topic] = {};

          try {
            const w = await wsclient;

            w.send(
              message({
                event: 'subscribe',
                data: topic,
              })
            );

            logger.log('subscribed to', topic);
          } catch (err) {
            logger.warn(err);
          }
        }
      })();

      const id = uuid();
      callbacks[topic][id] = callback;
      return id;
    },

    unsubscribe: (topic: string, id: string) => {
      (async () => {
        delete callbacks[topic][id];

        try {
          const w = await wsclient;

          if (Object.keys(callbacks[topic]).length === 0) {
            w.send(
              message({
                event: 'unsubscribe',
                data: topic,
              })
            );
          }
        } catch (err) {
          logger.warn(err);
        }
      })();
    },
  };
};

const SocketContext = createContext(socketContextDefaultValue);

const SocketProvider = ({ children }: ChildrenProps) => {
  const socket = useMemo(() => {
    if (typeof window !== 'undefined') {
      return createSocketContext();
    }

    return socketContextDefaultValue;
  }, [typeof window]);

  return (
    <SocketContext.Provider value={socket}>{children}</SocketContext.Provider>
  );
};

export const useSubscribe = (
  topics: string[],
  callback: (arg: { topic: string; message: string }) => void,
  dependencies: any[] = []
) => {
  const t = typeof topics === 'string' ? [topics] : topics;

  const { subscribe, unsubscribe } = useContext(SocketContext);

  useEffect(() => {
    const subscriptions = t.map((topic) => ({
      topic,
      id: subscribe(topic, callback),
    }));

    return () => {
      subscriptions.forEach(({ topic, id }) => unsubscribe(topic, id));
    };
  }, [...dependencies, ...t]);
};

export const useWatch = (...topics: string[]) => {
  const reloadPage = useReload();
  useSubscribe(
    topics,
    () => {
      console.log('hi');
      reloadPage();
    },
    topics
  );
};

export default SocketProvider;
