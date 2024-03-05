import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import * as wsock from 'websocket';
import { ChildrenProps } from '~/components/types';
import logger from '~/root/lib/client/helpers/log';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { socketUrl } from '~/root/lib/configs/base-url.cjs';

type IFor = 'logs' | 'resource-update';

export interface ISocketResp<T = any> {
  type: 'response' | 'error' | 'info';
  for: IFor;
  message: string;
  id: string;
  data: T;
}

type IData = {
  event?: 'subscribe' | 'unsubscribe';
  id: string;
};

interface ISocketMsg<T extends IData> {
  for: IFor;
  data: T;
}

interface IResponses {
  [key: string | IFor]: {
    [id: string]: ISocketResp[];
  };
}

type IsendMsg = <T extends IData>(msg: ISocketMsg<T>) => void;

const Context = createContext<{
  responses: IResponses;
  errors: IResponses;
  infos: IResponses;
  sendMsg: IsendMsg;
  clear: (msg: ISocketMsg<IData>) => void;
}>({
  clear: () => {},
  responses: {},
  errors: {},
  infos: {},
  sendMsg: () => {},
});

export const useSubscribe = <T extends IData>(
  msg: ISocketMsg<T> | ISocketMsg<T>[],
  dep: never[]
) => {
  const {
    sendMsg,
    responses,
    infos: i,
    errors: e,
    clear,
  } = useContext(Context);

  const [resp, setResp] = useState<ISocketResp[]>([]);
  const [subscribed, setSubscribed] = useState(false);
  const [errors, setErrors] = useState<ISocketResp[]>([]);
  const [infos, setInfos] = useState<ISocketResp[]>([]);

  useEffect(() => {
    (async () => {
      if (Array.isArray(msg)) {
        setResp(resp);

        const tr: ISocketResp[] = [];
        const terr: ISocketResp[] = [];
        const ti: ISocketResp[] = [];

        for (let k = 0; k < msg.length; k += 1) {
          const m = msg[k];

          tr.push(...(responses[m.for]?.[m.data.id || 'default'] || []));
          terr.push(...(e[m.for]?.[m.data.id || 'default'] || []));
          ti.push(...(i[m.for]?.[m.data.id || 'default'] || []));
        }
        setResp(tr);
        setErrors(terr);
        setInfos(ti);

        if (tr.length || ti.length) {
          setSubscribed(true);
        }
        return;
      }

      setResp(responses[msg.for]?.[msg.data.id || 'default'] || []);
      setErrors(e[msg.for]?.[msg.data.id || 'default'] || []);
      setInfos(i[msg.for]?.[msg.data.id || 'default'] || []);

      if (resp.length || i[msg.for]?.[msg.data.id || 'default']?.length) {
        setSubscribed(true);
      }
    })();
  }, [responses]);

  useDebounce(
    () => {
      console.log('subscribing');
      if (Array.isArray(msg)) {
        msg.forEach((m) => {
          sendMsg({ ...m, data: { ...m.data, event: 'subscribe' } });
        });
      } else {
        sendMsg({ ...msg, data: { ...msg.data, event: 'subscribe' } });
      }

      return () => {
        console.log('unsubscribing');
        if (Array.isArray(msg)) {
          msg.forEach((m) => {
            clear(m);
            setSubscribed(false);
            sendMsg({ ...m, data: { ...m.data, event: 'unsubscribe' } });
          });
          return;
        }

        clear(msg);
        setSubscribed(false);
        sendMsg({ ...msg, data: { ...msg.data, event: 'unsubscribe' } });
      };
    },
    1000,
    [...dep]
  );

  return {
    responses: resp,
    subscribed,
    infos,
    errors,
  };
};

export const SockProvider = ({ children }: ChildrenProps) => {
  const sockPromise = useRef<Promise<wsock.w3cwebsocket> | null>(null);

  const [responses, setResponses] = useState<IResponses>({});
  const [errors, setErrors] = useState<IResponses>({});
  const [infos, setInfos] = useState<IResponses>({});

  const setResp = useCallback((resp: ISocketResp) => {
    setResponses((s) => {
      const key = resp.for;
      const { id } = resp;
      return {
        ...s,
        [key]: {
          ...s[key],
          [id]: [...(s[key]?.[id] || []), resp],
        },
      };
    });
  }, []);

  const setError = useCallback((resp: ISocketResp) => {
    setErrors((s) => {
      const key = resp.for;
      const { id } = resp;
      return {
        ...s,
        [key]: {
          ...s[key],
          [id]: [...(s[key]?.[id] || []), resp],
        },
      };
    });
  }, []);

  const setInfo = useCallback((resp: ISocketResp) => {
    setInfos((s) => {
      const key = resp.for;
      const { id } = resp;
      return {
        ...s,
        [key]: {
          ...s[key],
          [id]: [...(s[key]?.[id] || []), resp],
        },
      };
    });
  }, []);

  const onMessage = useCallback((msg: wsock.IMessageEvent) => {
    try {
      const m: ISocketResp = JSON.parse(msg.data as string);

      switch (m.type) {
        case 'response':
          setResp(m);
          break;

        case 'info':
          setInfo(m);
          break;
        case 'error':
          setError(m);
          break;
        default:
          logger.log('unknown message', m);
      }
    } catch (err) {
      console.error(err);
    }
  }, []);

  useDebounce(
    () => {
      if (typeof window !== 'undefined') {
        try {
          sockPromise.current = new Promise<wsock.w3cwebsocket>((res, rej) => {
            let rejected = false;
            try {
              // eslint-disable-next-line new-cap
              const w = new wsock.w3cwebsocket(`${socketUrl}/ws`, '', '', {});

              w.onmessage = onMessage;

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
    },
    1000,
    []
  );

  const sendMsg = useCallback(
    async <T extends IData>(msg: ISocketMsg<T>) => {
      if (!sockPromise.current) {
        logger.log('no socket connection');
        return;
      }
      try {
        const w = await sockPromise.current;
        if (!w) {
          logger.log('no socket connection');
          return;
        }

        w.send(JSON.stringify(msg));
      } catch (err) {
        console.error(err);
      }
    },
    [sockPromise.current]
  );

  const clear = useCallback(
    <T extends IData>(msg: ISocketMsg<T>) => {
      setResponses((s) => {
        const key = msg.for;
        const id = msg.data.id || 'default';
        return {
          ...s,
          [key]: {
            ...s[key],
            [id]: [],
          },
        };
      });

      setErrors((s) => {
        const key = msg.for;
        const id = msg.data.id || 'default';
        return {
          ...s,
          [key]: {
            ...s[key],
            [id]: [],
          },
        };
      });

      setInfos((s) => {
        const key = msg.for;
        const id = msg.data.id || 'default';
        return {
          ...s,
          [key]: {
            ...s[key],
            [id]: [],
          },
        };
      });
    },
    [responses]
  );

  return (
    <Context.Provider
      value={useMemo(() => {
        return {
          clear,
          responses,
          errors,
          infos,
          sendMsg,
        };
      }, [responses, errors, infos, sendMsg])}
    >
      {children}
    </Context.Provider>
  );
};
