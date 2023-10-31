import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { INodepools } from '~/console/server/gql/queries/nodepool-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { IClusterContext } from '../_.$account.$cluster';
import HandleNodePool from './handle-nodepool';
import Resources from './resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { cluster } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listNodePools({
      clusterName: cluster,
      // pagination: getPagination(ctx),
      // search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { nodePoolData: data };
  });

  return defer({ promise });
};

const ClusterDetail = () => {
  const [showHandleNodePool, setHandleNodePool] =
    useState<IShowDialog<ExtractNodeType<INodepools> | null>>(null);

  const { promise } = useLoaderData<typeof loader>();

  const { cluster } = useOutletContext<IClusterContext>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ nodePoolData }) => {
          const nodepools = nodePoolData?.edges?.map(({ node }) => node);
          if (!nodepools) {
            return null;
          }

          console.log(nodepools);

          const { pageInfo, totalCount } = nodePoolData;
          return (
            <Wrapper
              header={{
                title: 'Nodepools',
                action: nodepools.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create new nodepool"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setHandleNodePool({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: nodepools.length === 0,
                title: 'This is where you’ll manage your cluster',
                content: (
                  <p>
                    You can create a new cluster and manage the listed cluster.
                  </p>
                ),
                action: {
                  content: 'Create new nodepool',
                  prefix: <Plus />,
                  LinkComponent: Link,
                  onClick: () => {
                    setHandleNodePool({ type: 'add', data: null });
                  },
                },
              }}
              pagination={{
                pageInfo,
                totalCount,
              }}
              tools={<Tools />}
            >
              <Resources
                items={nodepools}
                onEdit={(item) => {
                  setHandleNodePool({ type: 'edit', data: item });
                }}
              />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleNodePool
        show={showHandleNodePool}
        setShow={setHandleNodePool}
        cluster={cluster}
      />
    </>
  );
};

export default ClusterDetail;
