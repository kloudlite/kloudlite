import { useParams } from '@remix-run/react';
import {
  Dispatch,
  ReactNode,
  SetStateAction,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { useSocketWatch } from '~/root/lib/client/helpers/socket/useWatch';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useConsoleApi } from '../server/gql/api-provider';
import { ExtractNodeType, parseNodes } from '../server/r-utils/common';
import { IClustersStatus } from '../server/gql/queries/cluster-queries';

type IClusterMap = { [key: string]: ExtractNodeType<IClustersStatus> };

const ClusterStatusContext = createContext<{
  clusters: IClusterMap;
  setClusters: Dispatch<SetStateAction<IClusterMap>>;
}>({ clusters: {}, setClusters: () => {} });

const ClusterStatusProvider = ({ children }: { children: ReactNode }) => {
  const [clusters, setClusters] = useState<IClusterMap>({});
  const api = useConsoleApi();
  const [update, setUpdate] = useState(false);

  const { account } = useParams();

  const topic = useCallback(() => {
    return Object.keys(clusters).map((c) => `account:${account}.cluster:${c}`);
  }, [clusters])();

  const listCluster = useCallback(async () => {
    try {
      const cl = await api.listClusterStatus({
        pagination: {
          first: 5,
        },
      });
      const parsed = parseNodes(cl.data).reduce((acc, c) => {
        acc[c.metadata.name] = c;
        return acc;
      }, {} as IClusterMap);
      setClusters(parsed);
      return clusters;
    } catch (err) {
      console.error(err);
      return false;
    }
  }, []);

  useEffect(() => {
    const interval = setInterval(() => {
      listCluster();
    }, 30 * 1000);

    const onlineEvent = () => {
      setTimeout(() => {
        listCluster();
      }, 3000);
    };

    window.addEventListener('online', onlineEvent);

    return () => {
      clearInterval(interval);
      window.removeEventListener('online', onlineEvent);
    };
  }, []);

  useDebounce(
    () => {
      listCluster();
    },
    3000,
    [update],
  );

  useSocketWatch(() => {
    setUpdate((p) => !p);
  }, topic);

  return (
    <ClusterStatusContext.Provider
      value={useMemo(
        () => ({ clusters, setClusters }),
        [clusters, setClusters],
      )}
    >
      {children}
    </ClusterStatusContext.Provider>
  );
};

export default ClusterStatusProvider;

export const useClusterStatusV2 = () => {
  return useContext(ClusterStatusContext);
};
