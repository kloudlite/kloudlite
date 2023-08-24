import { useState } from 'react';
import { Link } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import ResourceList from '../../components/resource-list';
import { dummyData } from '../../dummy/data';
import Resources from './resources';
import Tools from './tools';
import HandleDomain from './handle-domain';

const ClusterDetail = () => {
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');
  const [showHandleNodePool, setHandleNodePool] = useState(null);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const [data, _setData] = useState(dummyData.domains);

  return (
    <>
      <Wrapper
        header={{
          title: 'Domains',
          action: data.length > 0 && (
            <Button
              variant="primary"
              content="Create new domain"
              prefix={PlusFill}
              onClick={() => {
                setHandleNodePool({ type: 'add', data: null });
              }}
            />
          ),
        }}
        empty={{
          is: data.length === 0,
          title: 'This is where you’ll oversees and control your domain.',
          content: (
            <p>
              You can add a new domain and exercise control over the domains
              listed.
            </p>
          ),
          action: {
            content: 'Create new domain',
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
        <Tools viewMode={viewMode} setViewMode={setViewMode} />
        <ResourceList mode={viewMode}>
          {data.map((cluster) => (
            <ResourceList.ResourceItem key={cluster.id} textValue={cluster.id}>
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

      {showHandleNodePool && (
        <HandleDomain show={showHandleNodePool} setShow={setHandleNodePool} />
      )}

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
