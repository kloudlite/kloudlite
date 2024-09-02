import { Link, useOutletContext, useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Badge } from '~/components/atoms/badge';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItem,
  ListItemV2,
  ListTitle,
  ListTitleV2,
  listClass,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import { Copy, GearSix, Pause, Play, Trash } from '~/console/components/icons';
import ListGridView from '~/console/components/list-grid-view';
import ListV2 from '~/console/components/listV2';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { findClusterStatus } from '~/console/hooks/use-cluster-status';
import { useClusterStatusV2 } from '~/console/hooks/use-cluster-status-v2';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IEnvironments } from '~/console/server/gql/queries/environment-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import CloneEnvironment from './clone-environment';

const RESOURCE_NAME = 'environment';
type BaseType = ExtractNodeType<IEnvironments>;

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

type OnAction = ({
  action,
  item,
}: {
  action: 'clone' | 'delete' | 'suspend' | 'resumed';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};

const ExtraButton = ({
  item,
  onAction,
  isClusterOnline,
}: IExtraButton & { isClusterOnline?: boolean }) => {
  const { account } = useParams();
  const iconSize = 16;
  let options: IResourceExtraItem[] = [
    {
      label: 'Clone',
      icon: <Copy size={iconSize} />,
      type: 'item',
      key: 'clone',
      onClick: () => onAction({ action: 'clone', item }),
    },
    {
      label: 'Delete',
      icon: <Trash size={iconSize} />,
      type: 'item',
      onClick: () => onAction({ action: 'delete', item }),
      key: 'delete',
      className: '!text-text-critical',
    },
  ];

  if (!item.isArchived) {
    if (isClusterOnline) {
      if (item.spec?.suspend) {
        options = [
          ...options,
          {
            label: 'Resume',
            icon: <Play size={iconSize} />,
            type: 'item',
            key: 'resumed',
            onClick: () => onAction({ action: 'resumed', item }),
          },
          {
            label: 'Settings',
            icon: <GearSix size={16} />,
            type: 'item',
            to: `/${account}/env/${parseName(item)}/settings/general`,
            key: 'settings',
          },
        ];
      } else {
        options = [
          ...options,
          {
            label: 'Suspend',
            icon: <Pause size={iconSize} />,
            type: 'item',
            key: 'suspend',
            onClick: () => onAction({ action: 'suspend', item }),
          },
          {
            label: 'Settings',
            icon: <GearSix size={16} />,
            type: 'item',
            to: `/${account}/env/${parseName(item)}/settings/general`,
            key: 'settings',
          },
        ];
      }
    } else {
      options = [
        ...options,
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',
          to: `/${account}/env/${parseName(item)}/settings/general`,
          key: 'settings',
        },
      ];
    }
  }

  return <ResourceExtraAction options={options} />;
};

interface IResource {
  items: (BaseType & { isClusterOnline: boolean })[];
  onAction: OnAction;
}

const GridView = ({ items = [], onAction }: IResource) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/env/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                className: listClass.title,
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton item={item} onAction={onAction} />}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: listClass.author,
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
  const { account } = useParams();
  const { clusters } = useClusterStatusV2();

  const [clusterOnlineStatus, setClusterOnlineStatus] = useState<
    Record<string, boolean>
  >({});
  useEffect(() => {
    const states: Record<string, boolean> = {};
    Object.entries(clusters).forEach(([key, value]) => {
      states[key] = findClusterStatus(value);
    });
    setClusterOnlineStatus(states);
  }, [clusters]);

  return (
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            render: () => 'Resource Name',
            name: 'name',
            className: listClass.title,
          },
          {
            render: () => 'Cluster',
            name: 'cluster',
            className: listClass.item,
          },
          {
            render: () => '',
            name: 'flex-post',
            className: listClass.flex,
          },
          {
            render: () => 'Status',
            name: 'status',
            className: listClass.status,
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
          const isClusterOnline = clusterOnlineStatus[i.clusterName];

          return {
            columns: {
              name: {
                render: () => (
                  <ListTitleV2
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              cluster: {
                render: () => (
                  <ListItemV2 data={i.isArchived ? '' : i.clusterName} />
                ),
              },
              status: {
                render: () => {
                  if (i.isArchived) {
                    return <Badge type="neutral">Archived</Badge>;
                  }

                  if (!isClusterOnline) {
                    return <Badge type="warning">Cluster Offline</Badge>;
                  }

                  if (i.spec?.suspend) {
                    return <Badge type="neutral">Suspended</Badge>;
                  }

                  return <SyncStatusV2 item={i} />;
                },
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
                    item={i}
                    onAction={onAction}
                    isClusterOnline={isClusterOnline}
                  />
                ),
              },
            },
            ...(i.isArchived ? {} : { to: `/${account}/env/${id}` }),
          };
        }),
      }}
    />
  );
};

const EnvironmentResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const { account } = useOutletContext<IAccountContext>();
  const api = useConsoleApi();
  const reloadPage = useReload();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.environment:${parseName(i)}`;
    })
  );

  const suspendEnvironment = async (item: BaseType, suspend: boolean) => {
    try {
      const { errors } = await api.updateEnvironment({
        env: {
          displayName: item.displayName,
          clusterName: item.clusterName,
          metadata: {
            name: parseName(item),
          },
          spec: {
            suspend,
          },
        },
      });

      if (errors) {
        throw errors[0];
      }
      toast.success(
        `${
          suspend
            ? 'Environment suspended successfully'
            : 'Environment resumed successfully'
        }`
      );
      reloadPage();
    } catch (err) {
      handleError(err);
    }
  };

  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [visible, setVisible] = useState<BaseType | null>(null);

  const props: IResource = {
    // @ts-ignore
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'clone':
          setVisible(item);
          break;
        case 'suspend':
          suspendEnvironment(item, true);
          break;
        case 'resumed':
          suspendEnvironment(item, false);
          break;
        case 'delete':
          setShowDeleteDialog(item);
          break;
        default:
          break;
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
        resourceName={parseName(showDeleteDialog)}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteEnvironment({
              envName: parseName(showDeleteDialog),
            });

            if (errors) {
              throw errors[0];
            }
            reloadPage();
            toast.success(`Environment deleted successfully`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
      <CloneEnvironment
        {...{
          isUpdate: true,
          visible: !!visible,
          setVisible: () => {
            setVisible(null);
          },
          data: visible!,
        }}
      />
    </>
  );
};

export default EnvironmentResourcesV2;
