import {
  Trash,
  PencilSimple,
  ArrowClockwise,
} from '~/iotconsole/components/icons';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/iotconsole/components/console-list-components';
import DeleteDialog from '~/iotconsole/components/delete-dialog';
import Grid from '~/iotconsole/components/grid';
import ListGridView from '~/iotconsole/components/list-grid-view';
import ResourceExtraAction from '~/iotconsole/components/resource-extra-action';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/iotconsole/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { IAccountContext } from '~/iotconsole/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { useOutletContext } from '@remix-run/react';
import ListV2 from '~/iotconsole/components/listV2';
// import { SyncStatusV2 } from '~/iotconsole/components/sync-status';
import { IBuilds } from '~/iotconsole/server/gql/queries/iot-build-queries';
import HandleBuild from './handle-builds';

type BaseType = ExtractNodeType<IBuilds>;
const RESOURCE_NAME = 'build';

const parseItem = (item: BaseType) => {
  return {
    name: item.name,
    id: item.id,
    status: item.status,
    cluster: item.buildClusterName,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IExtraButton {
  onDelete: () => void;
  onEdit: () => void;
  onTrigger: () => void;
}

const ExtraButton = ({ onDelete, onEdit, onTrigger }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Edit',
          icon: <PencilSimple size={16} />,
          type: 'item',
          onClick: onEdit,
          key: 'edit',
        },
        {
          label: 'Trigger',
          icon: <ArrowClockwise size={16} />,
          type: 'item',
          onClick: onTrigger,
          key: 'trigger',
        },
        {
          type: 'separator',
          key: 'separator1',
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
  onTrigger: (item: BaseType) => void;
}

const GridView = ({ items, onDelete, onEdit, onTrigger }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
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
                    action={
                      <ExtraButton
                        onDelete={() => {
                          onDelete(item);
                        }}
                        onEdit={() => {
                          onEdit(item);
                        }}
                        onTrigger={() => {
                          onTrigger(item);
                        }}
                      />
                    }
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

const ListView = ({ items, onDelete, onEdit, onTrigger }: IResource) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => 'Name',
            name: 'name',
            className: 'w-[180px] flex-1',
          },
          // {
          //   render: () => 'Status',
          //   name: 'status',
          //   className: 'flex-1 min-w-[30px] flex items-center justify-center',
          // },
          {
            render: () => 'Cluster',
            name: 'cluster',
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
          const { name, id, updateInfo, cluster } = parseItem(i);
          return {
            columns: {
              name: {
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              // status: {
              //   render: () =>
              //     i.latestBuildRun ? (
              //       <SyncStatusV2 item={i.latestBuildRun} />
              //     ) : null,
              // },
              cluster: { render: () => <ListItem data={cluster} /> },
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
                    onEdit={() => {
                      onEdit(i);
                    }}
                    onDelete={() => {
                      onDelete(i);
                    }}
                    onTrigger={() => {
                      onTrigger(i);
                    }}
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

const BuildResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleBuild, setHandleBuild] = useState<BaseType | null>(null);

  const api = useIotConsoleApi();
  const reloadPage = useReload();

  const { account } = useOutletContext<IAccountContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.id:${i.id}`;
    })
  );

  const triggerBuild = async (id: string) => {
    try {
      const { errors } = await api.triggerBuild({
        crTriggerBuildId: id,
      });

      if (errors) {
        throw errors[0];
      }
      reloadPage();
      toast.success(`${titleCase(RESOURCE_NAME)} triggered successfully`);
    } catch (err) {
      handleError(err);
    }
  };

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onEdit: (item) => {
      setHandleBuild(item);
    },
    onTrigger: async (item) => {
      await triggerBuild(item.id);
    },
  };

  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <DeleteDialog
        resourceName={showDeleteDialog?.name}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteBuild({
              crDeleteBuildId: showDeleteDialog?.id || '',
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
      <HandleBuild
        {...{
          isUpdate: true,
          data: showHandleBuild!,
          visible: !!showHandleBuild,
          setVisible: () => setHandleBuild(null),
        }}
      />
    </>
  );
};

export default BuildResources;
