import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useLoaderData, Link, useOutletContext } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { listOrGrid, parseName } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import ResourceList from '../../components/resource-list';
import HandleNodePool from './handle-nodepool';
import Resources from './resources';
import Tools from './tools';
import { IClusterContext } from '../_.$account.$cluster';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { cluster } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listNodePools({
      clusterName: cluster,
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { nodePoolData: data };
  });

  return defer({ promise });
};

const ClusterDetail = () => {
  const [viewMode, setViewMode] = useState<listOrGrid>('list');
  const [showHandleNodePool, setHandleNodePool] = useState<{
    data: any | null;
    type: 'add' | 'edit';
  } | null>(null);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

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
            >
              <Tools viewMode={viewMode} setViewMode={setViewMode} />
              <ResourceList mode={viewMode}>
                {nodepools.map((nodepool) => (
                  <ResourceList.ResourceItem
                    key={parseName(nodepool)}
                    textValue={parseName(nodepool)}
                  >
                    <Resources
                      item={nodepool}
                      onEdit={(e: any) => {
                        setHandleNodePool({ type: 'edit', data: e });
                      }}
                      onStop={(e: any) => {
                        setShowStopNodePool(e);
                      }}
                      onDelete={(e: any) => {
                        setShowDeleteNodePool(e);
                      }}
                    />
                  </ResourceList.ResourceItem>
                ))}
              </ResourceList>
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleNodePool
        show={showHandleNodePool}
        setShow={setHandleNodePool}
        cluster={cluster}
      />
      <AlertDialog
        show={showStopNodePool}
        setShow={setShowStopNodePool}
        title="Stop nodepool"
        message={"Are you sure you want to stop 'kloud-root-ca.crt'?"}
        type="warning"
        okText="Stop"
        onSubmit={(e) => {
          console.log(e);
        }}
      />
      <AlertDialog
        show={showDeleteNodePool}
        setShow={setShowDeleteNodePool}
        title="Delete nodepool"
        message={"Are you sure you want to delete 'kloud-root-ca.crt'?"}
        type="critical"
        okText="Delete"
        onSubmit={(e) => {
          console.log(e);
        }}
      />
    </>
  );
};

export default ClusterDetail;
