/* eslint-disable jsx-a11y/control-has-associated-label */
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext } from '@remix-run/react';
import { Box } from '~/console/components/common-console-components';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import Wrapper from '~/console/components/wrapper';
import {
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { HighlightJsLogs } from 'react-highlightjs-logs';
import { yamlDump } from '~/console/components/diff-viewer';
import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import { DownloadSimple } from '@jengaicons/react';
import { downloadFile, renderCloudProvider } from '~/console/utils/commons';
import { IClusterContext } from '../_.$account.$cluster';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listProviderSecrets({
      pagination: getPagination(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      providerSecrets: data,
    };
  });
  return defer({ promise });
};

const downloadConfig = ({
  value,
  encoding,
}: {
  value: string;
  encoding: 'base64' | string;
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
    filename: 'kubeconfig.yaml',
    data: linkSource,
    format: 'text/plain',
  });
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
  const { promise } = useLoaderData<typeof loader>();

  const { account, cluster } = useOutletContext<IClusterContext>();

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
    <LoadingComp data={promise}>
      {({ providerSecrets }) => {
        return (
          <Wrapper
            header={{
              title: 'Overview',
            }}
          >
            <div className="flex flex-col gap-6xl">
              <Box title="General">
                <div className="flex flex-col">
                  <div className="flex flex-row gap-3xl flex-wrap">
                    <ClusterInfoItem
                      title="Cluster name"
                      value={cluster.displayName}
                    />
                    <ClusterInfoItem
                      title="Cluster ID"
                      value={parseName(cluster)}
                    />
                    {!!cluster.adminKubeconfig && (
                      <ClusterInfoItem
                        title="Kube config"
                        value={
                          <Button
                            variant="primary-plain"
                            content="Download"
                            prefix={<DownloadSimple />}
                            onClick={() =>
                              downloadConfig(cluster.adminKubeconfig!)
                            }
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
                        cluster.spec?.cloudflareEnabled
                          ? 'Enabled'
                          : 'Disabled' || ''
                      }
                    />
                    <ClusterInfoItem
                      title="Backup To S3 Enabled"
                      value={
                        cluster.spec?.backupToS3Enabled
                          ? 'Enabled'
                          : 'Disabled' || ''
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
                      value={
                        cluster.spec?.taintMasterNodes ? 'true' : 'false' || ''
                      }
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
                <HighlightJsLogs
                  width="100%"
                  height="30rem"
                  //   title="Yaml Code"
                  dark
                  selectableLines
                  text={yamlDump(
                    cluster.status?.checks?.clusterApplyJob?.message || ''
                  )}
                  language="json"
                />
              </Box>
            </div>
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};
export default ClusterInfo;
