import { PencilLine, Trash } from '~/iotconsole/components/icons';
import { Link, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '@kloudlite/design-system/utils';
import ConsoleAvatar from '~/iotconsole/components/console-avatar';
import {
  ListItem,
  ListTitle,
} from '~/iotconsole/components/console-list-components';
import Grid from '~/iotconsole/components/grid';
import ListGridView from '~/iotconsole/components/list-grid-view';
import ResourceExtraAction from '~/iotconsole/components/resource-extra-action';
import {
  ExtractNodeType,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/iotconsole/server/r-utils/common';
import ListV2 from '~/iotconsole/components/listV2';
import { IDevices } from '~/iotconsole/server/gql/queries/iot-device-queries';
import { useState } from 'react';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { toast } from '@kloudlite/design-system/molecule/toast';
import DeleteDialog from '~/iotconsole/components/delete-dialog';
import { handleError } from '~/root/lib/utils/common';
import HandleDevice from './handle-device';

type BaseType = ExtractNodeType<IDevices>;
const RESOURCE_NAME = 'device';

const parseItem = (item: ExtractNodeType<IDevices>) => {
  return {
    name: item.displayName,
    id: item.name,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({
  onDelete,
  onEdit,
}: {
  onDelete: () => void;
  onEdit: () => void;
}) => {
  return (
    <ResourceExtraAction
      options={[
        {
          key: '1',
          label: 'Edit',
          icon: <PencilLine size={16} />,
          type: 'item',
          onClick: onEdit,
        },
        {
          label: 'Delete',
          icon: <Trash size={16} />,
          type: 'item',
          onClick: onDelete,
          key: 'delete',
          className: '!text-text-critical',
        },
      ]}
    />
  );
};

interface IResource {
  items: BaseType[];
  onDelete: (item: BaseType) => void;
  onEdit: (item: BaseType) => void;
}

const GridView = ({ items = [], onEdit, onDelete }: IResource) => {
  const { account, deployment } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${deployment}/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton
                        onDelete={() => onDelete(item)}
                        onEdit={() => onEdit(item)}
                      />
                    }
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
            ]}
          />
        );
      })}
    </Grid.Root>
  );
};

const ListView = ({ items, onEdit, onDelete }: IResource) => {
  return (
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            render: () => (
              <div className="flex flex-row">
                <span className="w-[48px]" />
                Name
              </div>
            ),
            name: 'name',
            className: 'w-[180px] flex-1',
          },
          {
            render: () => 'Updated',
            name: 'updated',
            className: 'w-[180px]',
          },
          {
            render: () => '',
            name: 'action',
            className: 'w-[24px]',
          },
        ],
        rows: items.map((i) => {
          const { name, id, updateInfo } = parseItem(i);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => (
                  <ExtraButton
                    onDelete={() => onDelete(i)}
                    onEdit={() => onEdit(i)}
                  />
                ),
              },
            },
          };
        }),
      }}
    />
  );
};

const DeviceResource = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleDevice, setShowHandleDevice] = useState<BaseType | null>(
    null
  );

  const api = useIotConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onEdit: (item) => {
      setShowHandleDevice(item);
    },
  };

  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <DeleteDialog
        resourceName={showDeleteDialog?.displayName}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteIotDevice({
              projectName: showDeleteDialog?.projectName || '',
              deploymentName: showDeleteDialog?.deploymentName || '',
              name: showDeleteDialog?.name || '',
            });

            if (errors) {
              throw errors[0];
            }
            reloadPage();
            toast.success(`${titleCase(RESOURCE_NAME)} deleted successfully`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
      <HandleDevice
        {...{
          isUpdate: true,
          visible: !!showHandleDevice,
          setVisible: () => {
            setShowHandleDevice(null);
          },
          data: showHandleDevice!,
        }}
      />
    </>
  );
};

export default DeviceResource;
