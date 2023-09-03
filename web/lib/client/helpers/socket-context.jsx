import { createContext, useContext, useEffect, useMemo } from 'react';
import { io } from 'socket.io-client';
import { v4 as uuid } from 'uuid';
import { socketUrl } from '~/root/lib/base-url';
import logger from './log';
import { useReload } from './reloader';

const createSocketContext = () => {
  const callbacks = {};
  const socket = io(socketUrl);

  socket.on('connect', () => {
    logger.log('socket connected');
  });

  socket.on('disconnect', () => {
    logger.log('socket disconnected');
  });

  socket.on('error', (error) => {
    logger.error(error);
  });

  return {
    subscribe: (topic, callback) => {
      if (!callbacks[topic]) {
        callbacks[topic] = {};
        // socket.emit('subscribe', topic);
        socket.on(`pubsub_${topic}`, (msg = '{}') => {
          // logger.log('message', topic);
          const { message } = JSON.parse(msg);
          logger.log('socket: ', topic);
          if (callbacks[topic]) {
            Object.values(callbacks[topic]).forEach((cb) =>
              cb({ topic, message })
            );
          }
        });
      }

      const id = uuid();
      callbacks[topic][id] = callback;
      return id;
    },

    unsubscribe: (topic, id) => {
      delete callbacks[topic][id];
      if (Object.keys(callbacks[topic]).length === 0) {
        // socket.emit('unsubscribe', topic);
        socket.off(`pubsub_${topic}`).removeAllListeners();
      }
    },
  };
};

const SocketContext = createContext(null);

const SocketProvider = ({ children }) => {
  const socket = useMemo(() => {
    // if (typeof window !== 'undefined') {
    //   return createSocketContext();
    // }
    //
    // return {
    //   subscribe: () => {},
    //   unsubscribe: () => {},
    // };
  }, [typeof window]);

  return (
    <SocketContext.Provider value={socket}>{children}</SocketContext.Provider>
  );
};

export const useSubscribe = (topics, callback, dependencies = []) => {
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

export const useWatch = (...topics) => {
  const reloadPage = useReload();
  useSubscribe(
    topics,
    () => {
      reloadPage();
    },
    topics
  );
};

export default SocketProvider;
