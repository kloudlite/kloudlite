import { PencilLine, QrCode, Trash } from '@jengaicons/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IShowDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDevices } from '~/console/server/gql/queries/vpn-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { useParams } from '@remix-run/react';
import HandleDevices, { ShowQR } from './handle-devices';

const RESOURCE_NAME = 'device';

const parseItem = (item: ExtractNodeType<IDevices>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    server: item.spec?.serverName,
    cluster: item.clusterName,
    updateInfo: {
      author: titleCase(
        `${parseUpdateOrCreatedBy(item)} updated the ${RESOURCE_NAME}`
      ),
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IExtraButton {
  onDelete: () => void;
  onQr: () => void;
  onEdit: () => void;
}
const ExtraButton = ({ onDelete, onQr, onEdit }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          icon: <PencilLine size={16} />,
          type: 'item',
          onClick: onEdit,
          key: 'edit',
        },
        {
          label: 'Show QR Code',
          icon: <QrCode size={16} />,
          type: 'item',
          onClick: onQr,
          key: 'qr',
        },
        {
          type: 'separator',
          key: 'sep-1',
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
  items: ExtractNodeType<IDevices>[];
  onDelete: (item: ExtractNodeType<IDevices>) => void;
  onQr: (item: ExtractNodeType<IDevices>) => void;
  onEdit: (item: ExtractNodeType<IDevices>) => void;
}

const GridView = ({
  items,
  onDelete = (_) => _,
  onQr = (_) => _,
  onEdit = (_) => _,
}: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, server, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitleWithSubtitle
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton
                        onDelete={() => {
                          onDelete(item);
                        }}
                        onQr={() => {
                          onQr(item);
                        }}
                        onEdit={() => {
                          onEdit(item);
                        }}
                      />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'access'),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={server} />
                    <ListBody data={cluster} />
                  </div>
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItemWithSubtitle
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

const ListView = ({
  items,
  onDelete = (_) => _,
  onQr = (_) => _,
  onEdit = (_) => _,
}: IResource) => {
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id, server, cluster, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitle title={name} subtitle={id} />
                ),
              },
              {
                key: generateKey(keyPrefix, 'server'),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={server} />,
              },
              {
                key: generateKey(keyPrefix, cluster),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={cluster} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItemWithSubtitle
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => (
                  <ExtraButton
                    onDelete={() => {
                      onDelete(item);
                    }}
                    onQr={() => {
                      onQr(item);
                    }}
                    onEdit={() => {
                      onEdit(item);
                    }}
                  />
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const DeviceResources = ({
  items = [],
}: {
  items: ExtractNodeType<IDevices>[];
}) => {
  const [showHandleDevice, setShowHandleDevice] =
    useState<IShowDialog<ExtractNodeType<IDevices> | null>>(null);
  const [showQR, setShowQR] = useState<IShowDialog<string>>(null);
  const [showDeleteDialog, setShowDeleteDialog] =
    useState<ExtractNodeType<IDevices> | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onQr: (item) => {
      setShowQR({ type: '', data: item.displayName });
    },
    onEdit: (item) => {
      setShowHandleDevice({ type: DIALOG_TYPE.EDIT, data: item });
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
      <ShowQR show={showQR} setShow={setShowQR} />
      <HandleDevices show={showHandleDevice} setShow={setShowHandleDevice} />
    </>
  );
};

export default DeviceResources;
