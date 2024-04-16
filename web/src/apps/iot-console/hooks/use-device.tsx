import { useEffect, useState } from 'react';
import axios from 'axios';
import { NN } from '~/root/lib/types/common';
import { useParams } from '@remix-run/react';
import { useConsoleApi } from '../server/gql/api-provider';
import { ExtractNodeType, parseNodes } from '../server/r-utils/common';
import { IDnsHosts } from '../server/gql/queries/cluster-queries';
import { IConsoleDevice } from '../server/gql/queries/console-vpn-queries';

type dev = {
  name: string;
  accountName: string;
  clusterName: string;
};

const getDevice = async (
  hosts: NN<ExtractNodeType<IDnsHosts>>[]
): Promise<dev> => {
  return new Promise((resolve, reject) => {
    const ref = {
      count: 0,
      done: false,
    };

    const iv = setInterval(() => {
      if (ref.done) {
        clearInterval(iv);
      }

      if (ref.count === hosts.length) {
        reject(new Error('No active device found'));
        clearInterval(iv);
        ref.done = true;
      }
    }, 1000);
    for (const host of hosts) {
      if (ref.done) {
        break;
      }

      axios({
        url: `https://whoami.vpn-device.${host.spec.publicDNSHost}:17172/whoami`,
        timeout: 2000,
      })
        .then((res) => {
          if (res.data.clusterName !== host.metadata.name) {
            reject(new Error('No active device found'));
            ref.done = true;
          }

          resolve(res.data);
        })
        .catch(() => {})
        .finally(() => {
          ref.count += 1;
        });
    }
  });
};

const useActiveDevice = () => {
  const api = useConsoleApi();
  const { account } = useParams();
  const [reload, setReload] = useState(true);
  const [state, setState] = useState<{
    device?: IConsoleDevice;
    loading: boolean;
    error?: Error;
  }>({
    loading: true,
  });
  useEffect(() => {
    (async () => {
      try {
        setState((prev) => ({ ...prev, loading: true }));
        const { data, errors } = await api.listDnsHosts();
        if (errors) {
          throw errors[0];
        }

        const hosts = parseNodes(data);

        const device = await getDevice(hosts);

        if (device.accountName !== account) {
          throw new Error('No active device found');
        }

        const { data: dev, errors: dErr } = await api.getConsoleVpnDevice({
          name: device.name,
        });

        if (dErr) {
          throw dErr[0];
        }

        setState((prev) => ({ ...prev, device: dev }));
      } catch (e) {
        const er = e as Error;
        setState((prev) => ({ ...prev, error: er }));
      } finally {
        setState((prev) => ({ ...prev, loading: false }));
      }
    })();
  }, [reload]);

  const reloadDevice = () => {
    setReload((s) => !s);
  };
  return { ...state, reloadDevice };
};

export default useActiveDevice;
