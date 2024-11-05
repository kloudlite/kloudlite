import { useCallback, useEffect, useState } from 'react';
import { useConsoleApi } from '../server/gql/api-provider';
import { parseNodes } from '../server/r-utils/common';

export const findClusterStatusv3 = (onlineAt?: string): boolean => {
  if (!onlineAt || !onlineAt) {
    return false;
  }

  const lastTime = new Date(onlineAt);
  const currentTime = new Date();

  const timeDifference =
    (currentTime.getTime() - lastTime.getTime()) / (1000 * 60);

  return timeDifference <= 1;
};

export const findClusterStatus = (item?: {
  lastOnlineAt?: string;
}): boolean => {
  if (!item || !item.lastOnlineAt) {
    return false;
  }

  const lastTime = new Date(item.lastOnlineAt);
  const currentTime = new Date();

  const timeDifference =
    (currentTime.getTime() - lastTime.getTime()) / (1000 * 60);

  return timeDifference <= 1;
};

const useClusterStatus = () => {
  const api = useConsoleApi();

  const [clusters, setClusters] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  const listCluster = useCallback(async () => {
    setLoading(true);
    try {
      const clusters = await api.listAllClusters();
      setClusters(parseNodes(clusters.data) || []);
      return clusters;
    } catch (err) {
      console.error(err);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    listCluster();
  }, []);

  return { findClusterStatus, clusters, loading };
};

export default useClusterStatus;
