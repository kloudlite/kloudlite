import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  useParams,
  useLoaderData,
  Link,
  useOutletContext,
} from '@remix-run/react';
import { defer } from '@remix-run/node';
import logger from '~/root/lib/client/helpers/log';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import {
  getPagination,
  getSearch,
  parseName,
} from '~/console/server/r-urils/common';
import ResourceList from '../../components/resource-list';
import HandleNodePool from './handle-nodepool';
import Resources from './resources';
import Tools from './tools';

const ClusterDetail = () => {
  const [viewMode, setViewMode] = useState('list');
  const [showHandleNodePool, setHandleNodePool] = useState(null);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const { account } = useParams();
  const { promise } = useLoaderData();

  // @ts-ignore
  const { cluster } = useOutletContext();

  return (
    <>
      <LoadingComp data={promise}>
        {({ nodePoolData }) => {
          const nodepools = nodePoolData?.edges?.map(({ node }) => node);
          if (!nodepools) {
            return null;
          }
          console.log(nodepools);
          const { pageInfo, totalCount } = nodepools;
          return (
            <Wrapper
              header={{
                title: 'Nodepools',
                action: nodepools.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create new nodepool"
                    prefix={PlusFill}
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
                  prefix: Plus,
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
                {nodepools.map((nodepool = {}) => (
                  <ResourceList.ResourceItem
                    key={parseName(nodepool)}
                    textValue={parseName(nodepool)}
                  >
                    <Resources
                      item={nodepool}
                      onEdit={(e) => {
                        setHandleNodePool({ type: 'edit', data: e });
                      }}
                      onStop={(e) => {
                        setShowStopNodePool(e);
                      }}
                      onDelete={(e) => {
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

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { cluster } = ctx.params;

  const promise = pWrapper(async () => {
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).listNodePools({
        clusterName: cluster,
        pagination: getPagination(ctx),
        search: getSearch(ctx),
      });
      if (errors) {
        throw errors[0];
      }
      return { nodePoolData: data };
    } catch (err) {
      logger.error(err);
      return { error: err.message };
    }
  });

  return defer({ promise });
};

export default ClusterDetail;
