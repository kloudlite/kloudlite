import { PencilLine, QrCode, Trash, WireGuardlogo } from '@jengaicons/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDevices } from '~/console/server/gql/queries/vpn-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { useParams } from '@remix-run/react';
import HandleDevices, { ShowWireguardConfig } from './handle-devices';

const RESOURCE_NAME = 'device';
type BaseType = ExtractNodeType<IDevices>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    cluster: item.clusterName,
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
      {items.map((item, index) => {
        const { name, id, cluster, updateInfo } = parseItem(item);
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
                key: generateKey(keyPrefix, 'access'),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={cluster} />
                  </div>
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
    <List.Root>
      {items.map((item, index) => {
        const { name, id, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'w-full',
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              {
                key: generateKey(keyPrefix, cluster),
                className: 'w-[180px] text-start mr-[50px]',
                render: () => <ListBody data={cluster} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[200px] min-w-[200px] max-w-[200px]',
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => <ExtraButton onAction={onAction} item={item} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const DeviceResources = ({ items = [] }: { items: BaseType[] }) => {
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

  const params = useParams();
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
            const { errors } = await api.deleteVpnDevice({
              deviceName: parseName(showDeleteDialog),
              clusterName: params.cluster || '',
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
      <HandleDevices
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

export default DeviceResources;
