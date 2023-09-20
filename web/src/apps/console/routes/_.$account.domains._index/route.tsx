import { Plus, PlusFill } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import Wip from '~/root/lib/client/components/wip';
import { dummyData } from '../../dummy/data';
import Tools from './tools';

const ClusterDetail = () => {
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [showHandleNodePool, setHandleNodePool] = useState<IShowDialog>(null);
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
              prefix={<PlusFill />}
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
            prefix: <Plus />,
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
        tools={<Tools />}
      >
        <Wip />
      </Wrapper>

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
