import { Plus } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { Link, useOutletContext } from '@remix-run/react';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import List from '~/console/components/list';
import ResourceList from '../../components/resource-list';
import { dummyData } from '../../dummy/data';
import HandleConfig from './handle-device';
import Resources from './resources';
import Tools from './tools';

const ProjectConfigIndex = () => {
  const [data, _setData] = useState(dummyData.projectConfig);
  const [showHandleConfig, setHandleConfig] = useState(false);
  const [showDeleteConfig, setShowDeleteConfig] = useState(false);

  const [_subNavAction, setSubNavAction] = useOutletContext();

  useEffect(() => {
    setSubNavAction({
      action: () => {
        setHandleConfig({ type: 'add', data: null });
      },
    });
  }, []);

  return (
    <>
      <Wrapper
        empty={{
          is: data.length === 0,
          title: 'This is where you’ll manage your Config.',
          content: (
            <p>You can create a new config and manage the listed configs.</p>
          ),
          action: {
            content: 'Create config',
            prefix: Plus,
            LinkComponent: Link,
            onClick: () => {
              setHandleConfig({ type: 'add', data: null });
            },
          },
        }}
      >
        <div className="flex flex-col">
          <Tools />
        </div>
        {/* <List /> */}
        <ResourceList mode="list" linkComponent={Link} prefetchLink>
          {data.map((d) => (
            <ResourceList.ResourceItem
              key={d.id}
              textValue={d.id}
              to={encodeURIComponent(d.name)}
            >
              <Resources
                item={d}
                onDelete={(item) => {
                  setShowDeleteConfig(item);
                }}
              />
            </ResourceList.ResourceItem>
          ))}
        </ResourceList>
      </Wrapper>
      <HandleConfig show={showHandleConfig} setShow={setHandleConfig} />
      {/* Alert Dialog for deleting config */}
      <AlertDialog
        show={showDeleteConfig}
        setShow={setShowDeleteConfig}
        title="Delete config"
        message={"Are you sure you want to delete 'kloud-root-ca.crt"}
        type="critical"
        okText="Delete"
        onSubmit={() => {}}
      />
    </>
  );
};

export default ProjectConfigIndex;

export const handle = {
  subheaderAction: () => <Button content="Add new config" prefix={Plus} />,
};
