import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useParams, useLoaderData, Link } from '@remix-run/react';
import { defer } from '@remix-run/node';
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
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');
  const [showHandleNodePool, setHandleNodePool] = useState(false);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const [data, _setData] = useState(dummyData.cluster);
  const { account } = useParams();
  const { promise } = useLoaderData();

  return (
    <>
      <LoadingComp data={promise}>
        {() => {
          return (
            <Wrapper
              header={{
                title: 'Cluster',
                backurl: `/${account}/clusters`,
                action: data.length > 0 && (
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
                is: data.length === 0,
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
                currentPage,
                itemsPerPage,
                totalItems,
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
                {data.map((cluster) => (
                  <ResourceList.ResourceItem
                    key={cluster.id}
                    textValue={cluster.id}
                  >
                    <Resources
                      item={cluster}
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

      <HandleNodePool show={showHandleNodePool} setShow={setHandleNodePool} />

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

export const loader = () => {
  const promise = pWrapper(async () => {
    return {};
  });
  return defer({ promise });
};

export default ClusterDetail;
