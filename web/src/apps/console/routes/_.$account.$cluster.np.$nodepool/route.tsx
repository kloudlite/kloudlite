/* eslint-disable jsx-a11y/control-has-associated-label */
import { defer } from '@remix-run/node';
import { useOutletContext, useParams } from '@remix-run/react';
import { Box } from '~/console/components/common-console-components';
import { pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import Wrapper from '~/console/components/wrapper';
import {
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import HighlightJsLog from '~/console/components/logger';
import { ReactNode } from 'react';
import { DownloadSimple } from '@jengaicons/react';
import { downloadFile, renderCloudProvider } from '~/console/utils/commons';
import { Chip } from '~/components/atoms/chips';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { IClusterContext } from '../_.$account.$cluster';

const ClusterTabs = () => {
  const { account, cluster } = useParams();
  return (
    <CommonTabs
      backButton={{
        to: `${account}/${cluster}/nodepools`,
        label: 'Nodepools',
      }}
    />
  );
};

export const handle = () => {
  return {
    navbar: <ClusterTabs />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { cluster, nodepool } = ctx.params;
    const { data, errors } = await GQLServerHandler(ctx.request).getNodePool({
      clusterName: cluster,
      poolName: nodepool,
    });

    if (errors) {
      throw errors[0];
    }

    return {
      nodepool: data,
    };
  });
  return defer({ promise });
};

const downloadConfig = ({
  value,
  encoding,
  filename,
}: {
  value: string;
  encoding: 'base64' | string;
  filename: string;
}) => {
  let linkSource = '';
  switch (encoding) {
    case 'base64':
      linkSource = atob(value);
      break;
    default:
      linkSource = value;
  }

  downloadFile({
    filename,
    data: linkSource,
    format: 'text/plain',
  });
};

const Log = () => {
  const getTime = () => {
    return Math.floor(new Date().getTime() / 1000);
  };

  const selectOptions = [
    {
      label: 'Last 12 hours',
      value: '1',
      from: () => getTime() - 43200,
    },
    {
      label: 'Last 24 hours',
      value: '2',
      from: () => getTime() - 86400,
    },
    {
      label: 'Last 7 days',
      value: '3',
      from: () => getTime() - 604800,
    },
    {
      label: 'Last 30 days',
      value: '3',
      from: () => getTime() - 2592000,
    },
  ];

  // const [from, setFrom] = useState(selectOptions[1].from());
  // const [selected, setSelected] = useState('1');

  const getUrl = (f: number) => {
    return `wss://observability.dev.kloudlite.io/observability/logs/pool-job?start_time=${f}&end_time=${getTime()}`;
  };

  // const [url, setUrl] = useState(getUrl(from));

  return (
    <HighlightJsLog
      // actionComponent={
      //   <Select
      //     size="md"
      //     options={async () => selectOptions}
      //     value={selectOptions[parseValue(selected, 1)]}
      //     onChange={(e) => {
      //       setSelected(e.value);
      //     }}
      //   />
      // }
      dark
      websocket
      height="60vh"
      width="100%"
      url={getUrl(selectOptions[3].from())}
      selectableLines
    />
  );
};

const ClusterInfoItem = ({
  title,
  value,
}: {
  title: ReactNode;
  value: ReactNode;
}) => {
  return (
    <div className="flex flex-col gap-lg flex-1 min-w-[45%]">
      <div className="bodyMd-medium text-text-default">{title}</div>
      <div className="bodyMd text-text-strong">{value}</div>
    </div>
  );
};
const ClusterInfo = () => {
  const { cluster } = useOutletContext<IClusterContext>();

  const providerInfo = () => {
    const provider = cluster.spec?.cloudProvider;
    switch (provider) {
      case 'aws':
        return (
          <ClusterInfoItem
            title="Region"
            value={cluster.spec?.aws?.region || ''}
          />
        );
      default:
        return null;
    }
  };
  return (
    <Wrapper
      header={{
        title: 'Cluster Info',
      }}
    >
      <div className="flex flex-col gap-6xl">
        <Box title="Overview">
          <div className="flex flex-col">
            <div className="flex flex-row gap-3xl flex-wrap">
              <ClusterInfoItem
                title="Cluster name"
                value={cluster.displayName}
              />
              <ClusterInfoItem title="Cluster ID" value={parseName(cluster)} />
              {!!cluster.adminKubeconfig && (
                <ClusterInfoItem
                  title="Kube config"
                  value={
                    <Chip
                      type="CLICKABLE"
                      item={cluster.adminKubeconfig}
                      label="Download"
                      prefix={<DownloadSimple />}
                      onClick={() => {
                        downloadConfig({
                          ...cluster.adminKubeconfig!,
                          filename: `${parseName(cluster)}-kubeconfig.yaml`,
                        });
                      }}
                    />
                  }
                />
              )}

              <ClusterInfoItem
                title="Last updated"
                value={`By ${parseUpdateOrCreatedBy(
                  cluster
                )} ${parseUpdateOrCreatedOn(cluster)}`}
              />
              <ClusterInfoItem
                title="Availability mode"
                value={cluster.spec?.availabilityMode || ''}
              />
              <ClusterInfoItem
                title="Cluster Internal Dns Host"
                value={cluster.spec?.clusterInternalDnsHost || ''}
              />
              <ClusterInfoItem
                title="Cloudflare Enabled"
                value={
                  cluster.spec?.cloudflareEnabled ? 'Enabled' : 'Disabled' || ''
                }
              />
              <ClusterInfoItem
                title="Backup To S3 Enabled"
                value={
                  cluster.spec?.backupToS3Enabled ? 'Enabled' : 'Disabled' || ''
                }
              />
              <ClusterInfoItem
                title="Kloudlite Release"
                value={cluster.spec?.kloudliteRelease || ''}
              />
              <ClusterInfoItem
                title="Public DNS Host"
                value={cluster.spec?.publicDNSHost || ''}
              />
              <ClusterInfoItem
                title="Taint Master Nodes"
                value={cluster.spec?.taintMasterNodes ? 'true' : 'false' || ''}
              />
              <ClusterInfoItem
                title="Cloud provider"
                value={renderCloudProvider({
                  cloudprovider: cluster.spec?.cloudProvider || 'unknown',
                })}
              />
              {providerInfo()}
            </div>
          </div>
        </Box>
        <Box title="Logs">
          <Log />
        </Box>
      </div>
    </Wrapper>
  );
};
export default ClusterInfo;
