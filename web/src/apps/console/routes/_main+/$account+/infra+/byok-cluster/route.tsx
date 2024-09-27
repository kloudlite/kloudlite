import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { Plus } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import fake from '~/root/fake-data-generator/fake';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import ByokClusterResource from './byok-cluster-resource';
import HandleByokCluster from './handle-byok-cluster';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);

    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listByokClusters({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      byokClusterData: data || {},
    };
  });

  return defer({ promise });
};

const ByocClusters = () => {
  // return <Wip />;
  const [visible, setVisible] = useState(false);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          byokClusterData: fake.ConsoleListByokClustersQuery
            .infra_listBYOKClusters as any,
        }}
      >
        {({ byokClusterData }) => {
          const byocClusterData = parseNodes(byokClusterData);

          return (
            <Wrapper
              secondaryHeader={{
                title: 'Clusters',
                action: byocClusterData.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Cluster"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: byocClusterData.length === 0,
                title:
                  'This is where you’ll attach your cluster or local devices.',
                content: (
                  <p>
                    You can create a new kubernetes cluster and manage the
                    listed kubernetes clusters or attach local devices.
                  </p>
                ),
                action: {
                  content: 'Create new Cluster',
                  prefix: <Plus />,
                  onClick: () => {
                    setVisible(true);
                  },
                  linkComponent: Link,
                },
              }}
              tools={<Tools />}
            // pagination={{
            //   pageInfo: byokClusterData.pageInfo,
            // }}
            >
              <ByokClusterResource items={byocClusterData} />
            </Wrapper>
          );
        }}
      </LoadingComp>
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

export default ByocClusters;
