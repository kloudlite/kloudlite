/* eslint-disable jsx-a11y/control-has-associated-label */
import { defer } from '@remix-run/node';
import { useOutletContext, useSearchParams } from '@remix-run/react';
import { Box } from '~/console/components/common-console-components';
import { pWrapper } from '~/console/components/loading-component';
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
import HighlightJsLog from '~/console/components/logger';
// import { HighlightJsLogs } from 'react-highlightjs-logs';
// import { yamlDump } from '~/console/components/diff-viewer';
import { ReactNode, useState } from 'react';
import { DownloadSimple } from '@jengaicons/react';
import { downloadFile, renderCloudProvider } from '~/console/utils/commons';
import { Chip } from '~/components/atoms/chips';
// import Wip from '~/console/components/wip';
import useForm from '~/root/lib/client/hooks/use-form';
import { useQueryParameters } from '~/root/lib/client/hooks/use-search';
import Yup from '~/root/lib/server/helpers/yup';
import { NumberInput } from '~/components/atoms/input';
import { Button } from '~/components/atoms/button';
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
  const [sp] = useSearchParams();

  // const url =
  //   'http://console-api.kl-core.svc.cluster.local:9100/observability/logs/cluster-job?start_time={{.startTime}}&end_time={{.endTime}}';

  const [url] = useState(
    `wss://observability.dev.kloudlite.io/observability/logs/cluster-job?start_time=${
      sp.get('start') || new Date().getTime() - 1000000
    }&end_time=${sp.get('end') || new Date().getTime()}`
  );

  const { setQueryParameters } = useQueryParameters();

  const { values, handleChange, handleSubmit } = useForm({
    initialValues: {
      start: sp.get('start') || new Date().getTime() - 1000000,
      end: sp.get('end') || new Date().getTime(),
    },
    validationSchema: Yup.object({}),
    onSubmit: (val) => {
      // @ts-ignore
      setQueryParameters(val);
    },
  });

  return (
    <div className="p-lg flex flex-col gap-xl">
      <div>Logs Url: {url}</div>
      <HighlightJsLog
        dark
        websocket
        height="60vh"
        width="100%"
        url={url}
        selectableLines
      />
      <form onSubmit={handleSubmit} className="flex flex-col gap-xl">
        <NumberInput
          label="start data timestamp"
          value={values.start}
          onChange={handleChange('start')}
          placeholder="start"
        />
        <NumberInput
          label="end date timestamp"
          placeholder="end"
          value={values.end}
          onChange={handleChange('end')}
        />
        <div className="flex gap-xl">
          <Button type="submit" content="update search params" />
          <Button
            onClick={() => {
              window.location.reload();
            }}
            content="reload"
          />
        </div>
      </form>
    </div>
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
