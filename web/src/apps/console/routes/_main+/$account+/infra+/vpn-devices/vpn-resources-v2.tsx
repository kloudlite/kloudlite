import { Eye, Trash } from '~/console/components/icons';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import { CopyButton, listRender } from '~/console/components/commons';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItem,
  ListItemV2,
  ListTitle,
  ListTitleV2,
  listClass,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
import { IGlobalVpnDevices } from '~/console/server/gql/queries/global-vpn-queries';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import DeleteDialog from '~/console/components/delete-dialog';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import { ShowWireguardConfig } from '~/console/page-components/handle-console-devices';

type BaseType = ExtractNodeType<IGlobalVpnDevices>;
const RESOURCE_NAME = 'global-vpn';

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({
  onDelete,
}: // onEdit,
{
  onDelete: () => void;
  // onEdit: () => void;
}) => {
  return (
    <ResourceExtraAction
      options={[
        // {
        //   key: '1',
        //   label: 'Edit',
        //   icon: <PencilLine size={16} />,
        //   type: 'item',
        //   onClick: onEdit,
        // },
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
  showWgConfig: (item: BaseType) => void;
  // onEdit: (item: BaseType) => void;
}

const DeviceHostView = ({ hostName }: { hostName: string }) => {
  return (
    <CopyContentToClipboard
      content={hostName}
      toastMessage="Device host copied successfully."
    />
  );
};

const GridView = ({ items = [], onDelete, showWgConfig }: IResource) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        const lR = listRender({ keyPrefix, resource: item });
        const status = lR.statusRender({ className: '' });
        return (
          <Grid.Column
            key={id}
            to={`/${account}/infra/${id}/overview`}
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
                        // onEdit={() => onEdit(item)}
                      />
                    }
                    // action={
                    //   // <ExtraButton status={status.status} cluster={item} />
                    //   <span />
                    // }
                  />
                ),
              },
              status,
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
const ListView = ({ items = [], onDelete, showWgConfig }: IResource) => {
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
            className: listClass.title,
          },
          {
            render: () => 'Device Config',
            name: 'config',
            className: listClass.item,
          },
          {
            render: () => 'Host',
            name: 'host',
            className: 'flex  w-[240px]',
          },
          {
            render: () => 'IP',
            name: 'ip',
            className: listClass.item,
          },
          {
            render: () => 'Updated',
            name: 'updated',
            className: listClass.updated,
          },
          {
            render: () => '',
            name: 'action',
            className: listClass.action,
          },
        ],
        rows: items.map((i) => {
          const { name, id, updateInfo } = parseItem(i);

          return {
            columns: {
              name: {
                render: () => (
                  <ListTitleV2
                    title={name || id}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              config: {
                render: () =>
                  i.creationMethod === '' ? (
                    <Button
                      variant="plain"
                      onClick={() => showWgConfig(i)}
                      content="View"
                      suffix={<Eye />}
                    />
                  ) : (
                    <Button
                      variant="plain"
                      onClick={() => showWgConfig(i)}
                      content="View"
                      suffix={<Eye />}
                    />
                  ),
              },
              host: {
                render: () => (
                  <div className="flex w-fit pulsable truncate">
                    <DeviceHostView hostName={`${parseName(i)}.device.local`} />
                  </div>
                ),
                // render: () => (
                //   <ListItem
                //     noTooltip
                //     data={
                //       <CopyButton
                //         title={
                //           <span className="text-sm">
                //             {parseName(i)}.device.local
                //           </span>
                //         }
                //         value={`${parseName(i)}.device.local`}
                //       />
                //     }
                //   />
                // ),
              },
              ip: {
                render: () => (
                  <ListItem
                    noTooltip
                    data={
                      <CopyButton
                        title={<span className="text-sm">{i.ipAddr}</span>}
                        value={i.ipAddr}
                      />
                    }
                  />
                ),
              },
              updated: {
                render: () => (
                  <ListItemV2
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => (
                  <ExtraButton
                    onDelete={() => onDelete(i)}
                    // onEdit={() => onEdit(i)}
                  />
                ),
              },
            },
            hideDetailSeperator: true,
          };
        }),
      }}
    />
  );
};

const VPNResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const { account } = useOutletContext<IAccountContext>();

  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showWireguardConfig, setShowWireguardConfig] =
    useState<BaseType | null>(null);

  const api = useConsoleApi();

  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    showWgConfig: (item) => {
      setShowWireguardConfig(item);
    },
    // onEdit: (item) => {
    //   setShowHandleGlobalVpnDevice(item);
    // },
  };

  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(i)}`;
    })
  );

  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <DeleteDialog
        // resourceName={showDeleteDialog?.displayName}
        resourceName={parseName(showDeleteDialog)}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteGlobalVpnDevice({
              gvpn: showDeleteDialog?.globalVPNName || '',
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
        setVisible={() => setShowWireguardConfig(null)}
        visible={!!showWireguardConfig}
        deviceName={parseName(showWireguardConfig)}
        creationMethod={showWireguardConfig?.creationMethod || ''}
      />
    </>
  );
};

export default VPNResourcesV2;
