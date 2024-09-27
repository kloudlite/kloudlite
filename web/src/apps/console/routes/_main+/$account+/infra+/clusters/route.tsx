import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { EmptyClusterImage } from '~/console/components/empty-resource-images';
import { Plus } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { IByocClusters } from '~/console/server/gql/queries/byok-cluster-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import fake from '~/root/fake-data-generator/fake';
import { IRemixCtx } from '~/root/lib/types/common';
import HandleByokCluster from '../byok-cluster/handle-byok-cluster';
import ClusterResourcesV2 from './cluster-resources-v2';
import Tools from './tools';

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
    return {
      clustersData: data,
    };
  });

  return defer({ promise });
};

const CreateClusterButton = () => {
  const [visible, setVisible] = useState(false);

  return (
    <>
      <Button
        content="Attach compute"
        variant="primary"
        prefix={<Plus />}
        onClick={() => {
          setVisible(true);
        }}
      />
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
  clustersData,
}: {
  clustersData: IByocClusters;
}) => {
  const [clusterType, setClusterType] = useState('All');
  const byokClusters = parseNodes(clustersData);

  const getEmptyState = ({
    byokClustersCount,
  }: {
    byokClustersCount: number;
  }) => {
    if (byokClustersCount > 0) {
      return {
        is: false,
        title: '',
        content: null,
        action: null,
      };
    }

    if (byokClustersCount === 0) {
      return {
        image: <EmptyClusterImage />,
        is: true,
        title: 'This is where you’ll attach your compute or local devices.',
        content: (
          <p>
            You can attach a new compute and manage the listed compute.
            <br />
            Follow the instructions to attach your{' '}
            <Link
              to="https://github.com/kloudlite/kl"
              className="text-text-default"
            >
              <span className="bodyMd-semibold underline underline-offset-1 text-text-default">
                local device
              </span>
            </Link>{' '}
          </p>
        ),
        action: <CreateClusterButton />,
      };
    }

    return {
      is: false,
      title: 'This is where you’ll manage your cluster.',
      content: (
        <p>You can create a new compute and manage the listed compute.</p>
      ),
      action: <CreateClusterButton />,
    };
  };

  if (!byokClusters) {
    return null;
  }
  return (
    <Wrapper
      secondaryHeader={{
        title: 'Attached Computes',
        action: byokClusters.length > 0 && <CreateClusterButton />,
      }}
      empty={getEmptyState({
        byokClustersCount: byokClusters.length,
      })}
      tools={
        <Tools
          onChange={(type) => {
            setClusterType(type);
          }}
          value={clusterType}
        />
      }
      pagination={clustersData}
    >
      <ClusterResourcesV2
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
        clustersData: fake.ConsoleListAllClustersQuery.byok_clusters as any,
      }}
    >
      {({ clustersData }) => {
        return <ClusterComponent clustersData={clustersData} />;
      }}
    </LoadingComp>
  );
};

export default Clusters;
