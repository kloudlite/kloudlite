/* eslint-disable jsx-a11y/control-has-associated-label */
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext, useParams } from '@remix-run/react';
import { Box } from '~/console/components/common-console-components';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterClientSide,
} from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import Wrapper from '~/console/components/wrapper';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import HighlightJsLog from '~/console/components/logger';
import { renderCloudProvider } from '~/console/utils/commons';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { DetailItem } from '~/console/components/commons';
import { INodepool } from '~/console/server/gql/queries/nodepool-queries';
import { IAccountContext } from '../_.$account';

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

const Log = ({ nodepool }: { nodepool: string }) => {
  const getTime = () => {
    return Math.floor(new Date().getTime() / 1000);
  };

  const { account } = useOutletContext<IAccountContext>();

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
  const params = useParams();
  ensureClusterClientSide(params);
  const getUrl = (f: number) => {
    return `wss://observability.dev.kloudlite.io/observability/logs/nodepool-job?resource_name=${nodepool}&resource_namespace=${
      account.spec.targetNamespace
    }&start_time=${f}&end_time=${getTime()}`;
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

const ClusterInfo = () => {
  const { promise } = useLoaderData<typeof loader>();

  const providerInfo = (nodepool: ExtractNodeType<INodepool>) => {
    const provider = nodepool.spec?.cloudProvider;
    switch (provider) {
      case 'aws':
        return (
          <DetailItem
            title="Availability zone"
            value={nodepool.spec?.aws?.availabilityZone || ''}
          />
        );
      default:
        return null;
    }
  };
  return (
    <LoadingComp data={promise}>
      {({ nodepool }) => {
        if (!nodepool) {
          return null;
        }
        return (
          <Wrapper
            header={{
              title: 'Nodepool Info',
            }}
          >
            <div className="flex flex-col gap-6xl">
              <Box title={`Nodepool Info (${nodepool.displayName})`}>
                <div className="flex flex-col">
                  <div className="flex flex-row gap-3xl flex-wrap">
                    <DetailItem
                      title="Nodepool ID"
                      value={parseName(nodepool)}
                    />

                    <DetailItem title="Cluster" value={nodepool.clusterName} />

                    <DetailItem
                      title="Last updated"
                      value={`By ${parseUpdateOrCreatedBy(
                        nodepool
                      )} ${parseUpdateOrCreatedOn(nodepool)}`}
                    />
                    <DetailItem
                      title="Cloud provider"
                      value={renderCloudProvider({
                        cloudprovider: nodepool.spec.cloudProvider || 'unknown',
                      })}
                    />
                    {providerInfo(nodepool)}
                  </div>
                </div>
              </Box>
              <Box title="Logs">
                <Log nodepool={parseName(nodepool)} />
              </Box>
            </div>
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};
export default ClusterInfo;
