import { useState } from 'react';
import { Link } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/AlertDialog';
import Wrapper from '~/console/components/Wrapper';
import ResourceList from '../../components/resource-list';
import { dummyData } from '../../dummy/data';
import HandleNodePool from './HandleNodePool';
import Filters from './Filters';
import Resource from './Resources';
import Tools from './Tools';

const ClusterDetail = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');
  const [showHandleNodePool, setHandleNodePool] = useState(false);
  const [nodePoolOperation, setNodePoolOperation] = useState('add');
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const [data, _setData] = useState(dummyData.cluster);

  return (
    <>
      <Wrapper
        header={{
          title: 'Cluster',
          backurl: '/bikash/clusters',
          action: data.length > 0 && (
            <Button
              variant="primary"
              content="Add nodepool"
              prefix={PlusFill}
              onClick={() => {
                setNodePoolOperation('add');
                setHandleNodePool(true);
              }}
            />
          ),
        }}
        empty={{
          is: data.length === 0,
          title: 'This is where you’ll manage your cluster',
          content: (
            <p>You can create a new cluster and manage the listed cluster.</p>
          ),
          action: {
            content: 'Create new cluster',
            prefix: Plus,
            LinkComponent: Link,
            onClick: () => {
              setNodePoolOperation('add');
              setHandleNodePool(true);
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
            <ResourceList.ResourceItem key={cluster.id} textValue={cluster.id}>
              <Resource
                item={cluster}
                onEdit={() => {
                  setNodePoolOperation('edit');
                  setHandleNodePool(true);
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

      <HandleNodePool
        show={showHandleNodePool}
        setShow={setHandleNodePool}
        type={nodePoolOperation}
        onSubmit={(values, errors, type) => {
          console.log(values, errors, type);
        }}
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
