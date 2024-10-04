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
import { useOutletContext, useParams } from '@remix-run/react';
import { useConsoleApi } from '../server/gql/api-provider';
import { parseName, parseNodes } from '../server/r-utils/common';
import { IAccountContext } from '../routes/_main+/$account+/_layout';

const ctx = createContext<{
  // clusters: {
  //   [key: string]: string;
  // };
  // setClusters: React.Dispatch<React.SetStateAction<{ [key: string]: string }>>;
  addToWatchList: (clusterNames: string[]) => void;
  removeFromWatchList: (clusterNames: string[]) => void;
}>({
  // clusters: {},
  // setClusters: () => {},
  addToWatchList: () => {},
  removeFromWatchList: () => {},
});

const ClusterStatusProvider = ({
  children,
  clustersMap,
  setClustersMap,
}: ChildrenProps & {
  clustersMap: { [key: string]: string };
  setClustersMap: React.Dispatch<
    React.SetStateAction<{ [key: string]: string }>
  >;
}) => {
  const [watchList, setWatchList] = useState<{
    [key: string]: number;
  }>({});

  const addToWatchList = (clusterNames: string[]) => {
    console.log('nayak', clusterNames);
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

    (async () => {
      try {
        const { data: clusters } = await api.listClusterStatus({
          pagination: {
            first: 100,
          },
          search: {
            allClusters: {
              exact: true,
              matchType: 'exact',
            },
            text: {
              array: keys,
              matchType: 'array',
            },
          },
        });

        const parsedNodes = parseNodes(clusters);

        const lastOnlineAt = parsedNodes.reduce((acc, curr) => {
          acc[parseName(curr)] = curr.lastOnlineAt;
          return acc;
        }, {} as { [key: string]: string });

        setClustersMap((s) => {
          return {
            ...s,
            ...lastOnlineAt,
          };
        });
      } catch (e) {
        console.log('error', e);
      }
    })();
  };

  useEffect(() => {
    const t = setInterval(() => {
      caller(watchList);
    }, 30 * 1000);

    return () => {
      clearInterval(t);
    };
  }, [watchList]);

  const { account } = useParams();

  const topic = useCallback(() => {
    return Object.keys(clustersMap).map(
      (c) => `account:${account}.cluster:${c}`
    );
  }, [clustersMap])();

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
          addToWatchList,
          removeFromWatchList,
        }),
        []
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
  const { clustersMap } = useOutletContext<IAccountContext>();
  const { addToWatchList, removeFromWatchList } = useContext(ctx);
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
    clustersMap,
  };
};
