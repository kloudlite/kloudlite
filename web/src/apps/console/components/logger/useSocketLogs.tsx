import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import * as wsock from 'websocket';
import { dayjs } from '~/components/molecule/dayjs';
import { ChildrenProps } from '~/components/types';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { socketUrl } from '~/root/lib/configs/base-url.cjs';

export type ILog = {
  pod_name: string;
  container_name: string;
  message: string;
  timestamp: string;
};

export type ISocketMessage = ILog;

const LogsContext = createContext<{
  sock: wsock.w3cwebsocket | null;
  logs: ISocketMessage[];
  resetLogs: () => void;
  subscribed: boolean;
  setSubscribed: (s: boolean) => void;
}>({
  sock: null,
  logs: [],
  resetLogs: () => {},
  subscribed: false,
  setSubscribed: () => {},
});

export interface IuseLog {
  url?: string;
  account: string;
  cluster: string;
  trackingId: string;
}

const useLogsContext = () => {
  return useContext(LogsContext);
};

export const LogsProvider = ({
  children,
  url,
}: ChildrenProps & {
  url?: string;
}) => {
  const [logs, setLogs] = useState<ISocketMessage[]>([]);

  const [sock, setSock] = useState<wsock.w3cwebsocket | null>(null);
  const sockPromise = useRef<Promise<wsock.w3cwebsocket> | null>(null);
  const [subscribed, setSubscribed] = useState(false);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      try {
        sockPromise.current = new Promise<wsock.w3cwebsocket>((res, rej) => {
          let rejected = false;
          try {
            // eslint-disable-next-line new-cap
            const w = new wsock.w3cwebsocket(
              url || `${socketUrl}/logs`,
              '',
              '',
              {}
            );

            w.onmessage = (msg) => {
              try {
                const m: {
                  timestamp: string;
                  message: string;
                  spec: {
                    podName: string;
                    containerName: string;
                  };
                  type: 'update' | 'error' | 'info';
                } = JSON.parse(msg.data as string);

                if (m.type === 'error') {
                  setLogs([]);
                  console.error(m.message);
                  return;
                }

                if (m.type === 'info') {
                  if (m.message === 'subscribed to logs') {
                    setSubscribed(true);
                  }
                  console.log(m.message);
                  return;
                }

                if (m.type === 'update') {
                  console.log(m.message);
                  return;
                }

                if (m.type === 'log') {
                  // setIsLoading(false);
                  setLogs((s) => [
                    ...s,
                    {
                      pod_name: m.spec.podName,
                      container_name: m.spec.containerName,
                      message: m.message,
                      timestamp: m.timestamp,
                    },
                  ]);
                  return;
                }

                console.log(m);
              } catch (err) {
                console.error(err);
              }
            };

            w.onopen = () => {
              res(w);
            };

            w.onerror = (e) => {
              console.error(e);
              if (!rejected) {
                rejected = true;
                rej(e);
              }
            };

            w.onclose = () => {};
          } catch (e) {
            rej(e);
          }
        });
      } catch (e) {
        console.log(e);
      }
    }
  }, []);

  useEffect(() => {
    const sorted = logs.sort((a, b) => {
      const resp = a.pod_name.localeCompare(b.pod_name);

      if (resp === 0) {
        return dayjs(a.timestamp).unix() - dayjs(b.timestamp).unix();
      }

      return resp;
    });

    if (JSON.stringify(sorted) !== JSON.stringify(logs)) {
      setLogs(sorted);
    }
  }, [logs]);

  useEffect(() => {
    (async () => {
      if (sockPromise.current) {
        const resp = await sockPromise.current;
        setSock(resp);
      }
    })();
  }, [sockPromise.current]);

  return (
    <LogsContext.Provider
      value={useMemo(() => {
        return {
          sock,
          logs,
          resetLogs: () => {
            setLogs([]);
          },
          subscribed,
          setSubscribed,
        };
      }, [sock, logs, subscribed])}
    >
      {children}
    </LogsContext.Provider>
  );
};

export const useSocketLogs = ({ account, cluster, trackingId }: IuseLog) => {
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  const { sock, subscribed, logs, resetLogs, setSubscribed } = useLogsContext();

  useDebounce(
    () => {
      try {
        if (account === '' || cluster === '' || trackingId === '') {
          return () => {};
        }
        if (logs.length) {
          resetLogs();
        }

        setIsLoading(true);

        sock?.send(
          JSON.stringify({
            event: 'subscribe',
            data: {
              account,
              cluster,
              trackingId,
            },
          })
        );
      } catch (e) {
        console.error(e);
        resetLogs();
        setError((e as Error).message);
      }

      return () => {
        setSubscribed(false);
        sock?.send(
          JSON.stringify({
            event: 'unsubscribe',
            data: {
              account,
              cluster,
              trackingId,
            },
          })
        );

        resetLogs();
      };
    },
    1000,
    [account, cluster, trackingId, sock]
  );

  return {
    logs,
    error,
    isLoading,
    subscribed,
  };
};
