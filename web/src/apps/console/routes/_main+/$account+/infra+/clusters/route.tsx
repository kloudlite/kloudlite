import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { ExtractNodeType, parseNodes } from '~/console/server/r-utils/common';
import {
  base64Encrypt,
  getPagination,
  getSearch,
} from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { useState } from 'react';
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import { IByocClusters } from '~/console/server/gql/queries/byok-cluster-queries';
import OptionList from '~/components/atoms/option-list';
import fake from '~/root/fake-data-generator/fake';
import { EmptyClusterImage } from '~/console/components/empty-resource-images';
import Tools from './tools';
import ClusterResourcesV2 from './cluster-resources-v2';
import HandleByokCluster from '../byok-cluster/handle-byok-cluster';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listAllClusters({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    if (!data.clusters?.totalCount) {
      const { data: secrets, errors: sErrors } = await GQLServerHandler(
        ctx.request
      ).listProviderSecrets({});

      if (sErrors) {
        throw sErrors[0];
      }

      return {
        clustersData: data || {},
        secretsCount: secrets.edges.length,
      };
    }
    return {
      clustersData: data,
      secretsCount: -1,
    };
  });

  return defer({ promise });
};

const CreateClusterButton = ({
  cpSecretsCount,
}: {
  cpSecretsCount?: number;
}) => {
  const { account } = useParams();

  const [visible, setVisible] = useState(false);

  return (
    <>
      <OptionList.Root>
        <OptionList.Trigger>
          <Button
            content={
              cpSecretsCount === 0
                ? 'setup cloud provider and cluster'
                : 'Add cluster'
            }
            variant="primary"
            prefix={<Plus />}
          />
        </OptionList.Trigger>
        <OptionList.Content>
          <OptionList.Link
            to={
              cpSecretsCount === 0
                ? `/onboarding/${account}/new-cloud-provider?f=${base64Encrypt(
                    'infra'
                  )}`
                : `/${account}/new-cluster`
            }
            LinkComponent={Link}
          >
            {cpSecretsCount === 0
              ? 'Setup cloud provider and cluster'
              : 'New Cluster'}
          </OptionList.Link>
          <OptionList.Item
            onClick={() => {
              setVisible(true);
            }}
          >
            Attach existing cluster
          </OptionList.Item>
        </OptionList.Content>
      </OptionList.Root>
      <HandleByokCluster
        {...{
          visible,
          setVisible,
          isUpdate: false,
        }}
      />
    </>
  );
};

const ClusterComponent = ({
  clusters,
  byokClusters,
  secretsCount,
}: // pageInfo,
// totalCount,
{
  clusters: ExtractNodeType<IClusters>[];
  byokClusters: ExtractNodeType<IByocClusters>[];
  secretsCount: number;
  // pageInfo: IClusters['pageInfo'];
  // totalCount: IClusters['totalCount'];
}) => {
  const [clusterType, setClusterType] = useState('All');

  const getEmptyState = ({
    clustersCount,
    byokClustersCount,
    cloudProviderSecretsCount,
  }: {
    clustersCount: number;
    byokClustersCount: number;
    cloudProviderSecretsCount: number;
  }) => {
    if (byokClustersCount > 0 || clustersCount > 0) {
      return {
        is: false,
        title: '',
        content: null,
        action: null,
      };
    }

    if (cloudProviderSecretsCount === 0) {
      return {
        image: <EmptyClusterImage />,
        is: true,
        title:
          'please setup your cloud provider first or attach your own cluster',
        content: (
          <p>
            you need to setup your add at least one cloud provider first or
            attch your own cluster, before starting working with clusters
          </p>
        ),
        action: (
          <CreateClusterButton cpSecretsCount={cloudProviderSecretsCount} />
        ),
      };
    }

    if (clustersCount === 0 && byokClustersCount === 0) {
      return {
        image: <EmptyClusterImage />,
        is: true,
        title: 'This is where you’ll manage your cluster.',
        content: (
          <p>You can create a new cluster and manage the listed cluster.</p>
        ),
        action: <CreateClusterButton />,
      };
    }

    return {
      is: false,
      title: 'This is where you’ll manage your cluster.',
      content: (
        <p>You can create a new cluster and manage the listed cluster.</p>
      ),
      action: <CreateClusterButton />,
    };
  };

  if (!clusters || !byokClusters) {
    return null;
  }

  return (
    <Wrapper
      secondaryHeader={{
        title: 'Clusters',
        action: (clusters.length > 0 || byokClusters.length > 0) && (
          <CreateClusterButton cpSecretsCount={secretsCount} />
        ),
      }}
      empty={getEmptyState({
        clustersCount: clusters.length,
        byokClustersCount: byokClusters.length,
        cloudProviderSecretsCount: secretsCount,
      })}
      tools={
        <Tools
          onChange={(type) => {
            setClusterType(type);
          }}
          value={clusterType}
        />
      }
    >
      <ClusterResourcesV2
        items={clusterType !== 'Byok' ? clusters : []}
        byokItems={clusterType !== 'Normal' ? byokClusters : []}
      />
    </Wrapper>
  );
};

const Clusters = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        clustersData: fake.ConsoleListAllClustersQuery as any,
        secretsCount: 1,
      }}
    >
      {({ clustersData, secretsCount }) => {
        const clusters = parseNodes(clustersData.clusters);

        const byokClusters = parseNodes(clustersData.byok_clusters);

        return (
          <ClusterComponent
            clusters={clusters}
            byokClusters={byokClusters}
            secretsCount={secretsCount}
          />
        );
      }}
    </LoadingComp>
  );
};

export default Clusters;
