import { Plus, PlusFill } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import Wip from '~/console/components/wip';
import Wrapper from '~/console/components/wrapper';
import { dummyData } from '../../dummy/data';
import Tools from './tools';

const ClusterDetail = () => {
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);

  const [data, _setData] = useState(dummyData.domains);

  return (
    <Wrapper
      header={{
        title: 'Domains',
        action: data.length > 0 && (
          <Button
            variant="primary"
            content="Create new domain"
            prefix={<PlusFill />}
            onClick={() => {
              // setHandleNodePool({ type: 'add', data: null });
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
            // setHandleNodePool({ type: 'add', data: null });
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
  );
};
export default ClusterDetail;
