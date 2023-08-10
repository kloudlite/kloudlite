import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useParams, useLoaderData, Link } from '@remix-run/react';
import { defer } from '@remix-run/node';
import logger from '~/root/lib/client/helpers/log';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import {
  getPagination,
  getSearch,
  parseName,
} from '~/console/server/r-urils/common';
import ResourceList from '../../components/resource-list';
import { dummyData } from '../../dummy/data';
import HandleNodePool from './handle-nodepool';
import Filters from './filters';
import Resources from './resources';
import Tools from './tools';

const ClusterDetail = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [viewMode, setViewMode] = useState('list');
  const [showHandleNodePool, setHandleNodePool] = useState(false);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const { account } = useParams();
  const { promise, clusterPromise } = useLoaderData();

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
                title: 'Cluster',
                backurl: `/${account}/clusters`,
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
              <div className="flex flex-col">
                <Tools viewMode={viewMode} setViewMode={setViewMode} />
                <Filters
                  appliedFilters={appliedFilters}
                  setAppliedFilters={setAppliedFilters}
                />
              </div>
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

      <LoadingComp data={clusterPromise} skeleton={<span />}>
        {({ cluster }) => {
          return (
            <HandleNodePool
              show={showHandleNodePool}
              setShow={setHandleNodePool}
              cluster={cluster}
            />
          );
        }}
      </LoadingComp>

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
  const { cluster, account } = ctx.params;

  const clusterPromise = pWrapper(async () => {
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).getCluster({
        name: cluster,
      });
      if (errors) {
        throw errors[0];
      }
      return { cluster: data };
    } catch (err) {
      logger.error(err);
      return { redirect: `/${account}/clusters` };
    }
  });

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

  return defer({ promise, clusterPromise });
};

export default ClusterDetail;
