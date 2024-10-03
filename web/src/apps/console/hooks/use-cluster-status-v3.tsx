import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { ChildrenProps } from '~/components/types';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useSocketWatch } from '~/root/lib/client/helpers/socket/useWatch';
import { useParams } from '@remix-run/react';
import { useConsoleApi } from '../server/gql/api-provider';

const ctx = createContext<{
  clusters: {
    [key: string]: string;
  };
  setClusters: React.Dispatch<React.SetStateAction<{ [key: string]: string }>>;
  addToWatchList: (clusterNames: string[]) => void;
  removeFromWatchList: (clusterNames: string[]) => void;
}>({
  clusters: {},
  setClusters: () => {},
  addToWatchList: () => {},
  removeFromWatchList: () => {},
});

const ClusterStatusProvider = ({ children }: ChildrenProps) => {
  const [clusters, setClusters] = useState<{
    [key: string]: string;
  }>({});
  const [watchList, setWatchList] = useState<{
    [key: string]: number;
  }>({});

  const addToWatchList = (clusterNames: string[]) => {
    setWatchList((s) => {
      const resp = clusterNames.reduce((acc, curr) => {
        if (!curr) {
          return acc;
        }
        if (acc[curr]) {
          acc[curr] += acc[curr];
        } else {
          acc[curr] = 1;
        }

        return acc;
      }, s);

      return resp;
    });
  };

  const api = useConsoleApi();

  const caller = (wl: { [key: string]: number }) => {
    const keys = Object.keys(wl);
    // console.log('nayak', wl, keys, Object.entries(wl));
    for (let i = 0; i < keys.length; i += 1) {
      (async () => {
        const w = keys[i];
        try {
          const { data: cluster } = await api.getClusterStatus({
            name: w,
          });
          setClusters((s) => {
            return {
              ...s,
              [w]: cluster.lastOnlineAt,
            };
          });
        } catch (e) {
          console.log('error', e);
        }
      })();
    }
  };

  useEffect(() => {
    const t2 = setTimeout(() => {
      caller(watchList);
    }, 1000);

    const t = setInterval(() => {
      caller(watchList);
    }, 30 * 1000);

    return () => {
      clearTimeout(t2);
      clearInterval(t);
    };
  }, [watchList]);

  const { account } = useParams();

  const topic = useCallback(() => {
    return Object.keys(clusters).map((c) => `account:${account}.cluster:${c}`);
  }, [clusters])();

  useSocketWatch(() => {
    caller(watchList);
  }, topic);

  const removeFromWatchList = (clusterNames: string[]) => {
    setWatchList((s) => {
      const resp = clusterNames.reduce((acc, curr) => {
        if (!curr) {
          return acc;
        }

        if (acc[curr] && acc[curr] >= 1) {
          acc[curr] -= acc[curr];
        }

        if (acc[curr] === 0) {
          delete acc[curr];
        }

        return acc;
      }, s);

      return resp;
    });
  };
  return (
    <ctx.Provider
      value={useMemo(
        () => ({
          clusters,
          setClusters,
          addToWatchList,
          removeFromWatchList,
        }),
        [clusters, setClusters]
      )}
    >
      {children}
    </ctx.Provider>
  );
};

export default ClusterStatusProvider;

export const useClusterStatusV3 = ({
  clusterName,
  clusterNames,
}: {
  clusterName?: string;
  clusterNames?: string[];
}) => {
  const { clusters, addToWatchList, removeFromWatchList } = useContext(ctx);
  useDebounce(
    () => {
      if (!clusterName && !clusterNames) {
        return () => {};
      }

      if (clusterName) {
        addToWatchList([clusterName]);
      } else if (clusterNames) {
        addToWatchList(clusterNames);
      }

      return () => {
        if (clusterName) {
          removeFromWatchList([clusterName]);
        } else if (clusterNames) {
          removeFromWatchList(clusterNames);
        }
      };
    },
    100,
    [clusterName, clusterNames]
  );

  return {
    clusters,
  };
};
