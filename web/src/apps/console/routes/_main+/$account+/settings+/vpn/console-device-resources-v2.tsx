import { PencilLine, QrCode, Trash, WireGuardlogo } from '@jengaicons/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/lib/client/helpers/reloader';
import { handleError } from '~/lib/utils/common';
import { IConsoleDevices } from '~/console/server/gql/queries/console-vpn-queries';
import HandleConsoleDevices, {
  ShowWireguardConfig,
} from '~/console/page-components/handle-console-devices';
import ListV2 from '~/console/components/listV2';
import ConsoleAvatar from '~/console/components/console-avatar';
import { SyncStatusV2 } from '~/console/components/sync-status';

const RESOURCE_NAME = 'device';
type BaseType = ExtractNodeType<IConsoleDevices>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: titleCase(
        `${parseUpdateOrCreatedBy(item)} updated the ${RESOURCE_NAME}`
      ),
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'edit' | 'delete' | 'qr' | 'config';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};
const ExtraButton = ({ onAction, item }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          icon: <PencilLine size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'edit', item }),
          key: 'edit',
        },
        {
          label: 'Show QR Code',
          icon: <QrCode size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'qr', item }),
          key: 'qr',
        },
        {
          label: 'Show Wireguard Config',
          icon: <WireGuardlogo size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'config', item }),
          key: 'wireguard-config',
        },
        {
          type: 'separator',
          key: 'sep-1',
        },
        {
          label: 'Delete',
          icon: <Trash size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'delete', item }),
          key: 'delete',
          className: '!text-text-critical',
        },
      ]}
    />
  );
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items, onAction }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items?.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton onAction={onAction} item={item} />}
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

const ListView = ({ items, onAction }: IResource) => {
  return (
    <ListV2.Root
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
            className: 'w-[180px]',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
          },
          {
            render: () => 'Project',
            name: 'project',
            className: 'w-[180px]',
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
              status: {
                render: () => <SyncStatusV2 item={i} />,
              },
              project: { render: () => <ListItem data={i.projectName} /> },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => <ExtraButton onAction={onAction} item={i} />,
              },
            },
          };
        }),
      }}
    />
  );
};

const ConsoleDeviceResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const [showHandleDevice, setShowHandleDevice] = useState<BaseType | null>(
    null
  );
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showWireguardConfig, setShowWireguardConfig] = useState<{
    device: string;
    mode: 'qr' | 'config';
  } | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'edit':
          setShowHandleDevice(item);
          break;
        case 'delete':
          setShowDeleteDialog(item);
          break;
        case 'qr':
          setShowWireguardConfig({ device: parseName(item), mode: 'qr' });
          break;
        case 'config':
          setShowWireguardConfig({ device: parseName(item), mode: 'config' });
          break;
        default:
      }
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
            const { errors } = await api.deleteConsoleVpnDevice({
              deviceName: parseName(showDeleteDialog),
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
      <ShowWireguardConfig
        {...{
          visible: !!showWireguardConfig,
          setVisible: () => setShowWireguardConfig(null),
          data: showWireguardConfig!,
          mode: showWireguardConfig?.mode || 'config',
        }}
      />
      <HandleConsoleDevices
        {...{
          isUpdate: true,
          data: showHandleDevice!,
          visible: !!showHandleDevice,
          setVisible: () => setShowHandleDevice(null),
        }}
      />
    </>
  );
};

export default ConsoleDeviceResourcesV2;
