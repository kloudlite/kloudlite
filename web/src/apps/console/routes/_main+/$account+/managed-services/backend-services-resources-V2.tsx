import { Link, useOutletContext, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Badge } from '~/components/atoms/badge';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListItemV2,
  ListTitle,
  ListTitleV2,
  listClass,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import { GearSix, Trash } from '~/console/components/icons';
import ListGridView from '~/console/components/list-grid-view';
import ListV2 from '~/console/components/listV2';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { findClusterStatusv3 } from '~/console/hooks/use-cluster-status';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IClusterMSvs } from '~/console/server/gql/queries/cluster-managed-services-queries';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { getManagedTemplate } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useWatchReload } from '~/root/lib/client/helpers/socket/useWatch';
import { handleError } from '~/root/lib/utils/common';
import { useClusterStatusV3 } from '~/console/hooks/use-cluster-status-v3';
import { IAccountContext } from '../_layout';
import { IClusterContext } from '../infra+/$cluster+/_layout';
import CloneManagedService from './clone-managed-service';

const RESOURCE_NAME = 'managed service';
type BaseType = ExtractNodeType<IClusterMSvs>;

const parseItem = (item: BaseType, templates: IMSvTemplates) => {
  const template = getManagedTemplate({
    templates,
    kind: item.spec?.msvcSpec?.serviceTemplate.kind || '',
    apiVersion: item.spec?.msvcSpec?.serviceTemplate.apiVersion || '',
  });
  return {
    name: item?.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
    logo: template?.logoUrl,
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'clone' | 'delete';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};

const ExtraButton = ({ item, onAction }: IExtraButton) => {
  const { account } = useParams();
  return item.isArchived ? (
    <ResourceExtraAction
      options={[
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
  ) : (
    <ResourceExtraAction
      options={[
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',

          to: `/${account}/msvc/${parseName(item)}/settings`,
          key: 'settings',
        },
      ]}
    />
  );
};

interface IResource {
  items: BaseType[];
  templates: IMSvTemplates;
  onAction: OnAction;
}

const GridView = ({ items, templates, onAction }: IResource) => {
  const { account, project } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, logo, updateInfo } = parseItem(item, templates);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${project}/msvc/${id}/logs-n-metrics`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton item={item} onAction={onAction} />
                      // <ResourceExtraAction
                      //   options={[
                      //     {
                      //       key: 'managed-services-resource-extra-action-1',
                      //       to: `/${account}/${project}/msvc/${id}/logs-n-metrics`,
                      //       icon: <GearSix size={16} />,
                      //       label: 'logs & metrics',
                      //       type: 'item',
                      //     },
                      //   ]}
                      // />
                    }
                    avatar={
                      <img src={logo} alt={name} className="w-4xl h-4xl" />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'author'),
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

const ListView = ({ items, templates, onAction }: IResource) => {
  const { account } = useOutletContext<IAccountContext>();
  const { clustersMap: clusterStatus } = useClusterStatusV3({
    clusterNames: items.map((i) => i.clusterName),
  });

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
          const isClusterOnline = findClusterStatusv3(
            clusterStatus[i.clusterName]
          );
          const { name, id, logo, updateInfo } = parseItem(i, templates);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitleV2
                    title={name}
                    subtitle={id}
                    avatar={
                      <div className="pulsable pulsable-circle aspect-square">
                        <img src={logo} alt={name} className="w-4xl h-4xl" />
                      </div>
                    }
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

                  if (clusterStatus[i.clusterName] === undefined) {
                    return <ListItemV2 className="px-4xl" data="-" />;
                  }

                  if (!isClusterOnline) {
                    return <Badge type="warning">Cluster Offline</Badge>;
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
                render: () => <ExtraButton item={i} onAction={onAction} />,
              },
            },
            ...(i.isArchived
              ? {}
              : { to: `/${parseName(account)}/msvc/${id}/managed-resources` }),
          };
        }),
      }}
    />
  );
};

const BackendServicesResourcesV2 = ({
  items = [],
  templates = [],
}: {
  items: BaseType[];
  templates: IMSvTemplates;
}) => {
  const { account } = useOutletContext<IClusterContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.cluster:${
        i.clusterName
      }.cluster_managed_service:${parseName(i)}`;
    })
  );

  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [visible, setVisible] = useState<BaseType | null>(null);
  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    templates,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'clone':
          setVisible(item);
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
            const { errors } = await api.deleteClusterMSv({
              name: parseName(showDeleteDialog),
            });

            if (errors) {
              throw errors[0];
            }
            reloadPage();
            toast.success(`Integrated service deleted successfully`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
      <CloneManagedService
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

export default BackendServicesResourcesV2;
