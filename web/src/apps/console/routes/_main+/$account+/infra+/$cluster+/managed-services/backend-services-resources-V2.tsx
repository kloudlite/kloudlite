import { GearSix, PencilSimple, Trash } from '~/console/components/icons';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import { getManagedTemplate } from '~/console/utils/commons';
import DeleteDialog from '~/console/components/delete-dialog';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useState } from 'react';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
import { IClusterMSvs } from '~/console/server/gql/queries/cluster-managed-services-queries';
import HandleBackendService from './handle-backend-service';
import { IClusterContext } from '../_layout';

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
  action: 'delete' | 'edit';
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
          icon: <PencilSimple size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'edit', item }),
          key: 'edit',
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
  templates: IMSvTemplates;
  onAction: OnAction;
}

const GridView = ({ items = [], templates = [], onAction: _ }: IResource) => {
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
                      <ResourceExtraAction
                        options={[
                          {
                            key: 'managed-services-resource-extra-action-1',
                            to: `/${account}/${project}/msvc/${id}/logs-n-metrics`,
                            icon: <GearSix size={16} />,
                            label: 'logs & metrics',
                            type: 'item',
                          },
                        ]}
                      />
                    }
                    // action={<ExtraButton onAction={onAction} item={item} />}
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

const ListView = ({ items = [], templates = [], onAction }: IResource) => {
  const { account, cluster } = useOutletContext<IClusterContext>();
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
            className: 'w-[180px]',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
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
          const { name, id, logo, updateInfo } = parseItem(i, templates);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
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
              status: {
                render: () => <SyncStatusV2 item={i} />,
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
                render: () => <ExtraButton item={i} onAction={onAction} />,
              },
            },
            to: `/${parseName(account)}/infra/${parseName(
              cluster
            )}/msvc/${id}/logs-n-metrics`,
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
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [visible, setVisible] = useState<BaseType | null>(null);
  const api = useConsoleApi();
  const reloadPage = useReload();
  const params = useParams();

  const { account, cluster } = useOutletContext<IClusterContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(
        cluster
      )}.managed_service:${parseName(i)}`;
    })
  );

  const props: IResource = {
    items,
    templates,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'delete':
          setShowDeleteDialog(item);
          break;
        case 'edit':
          setVisible(item);
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
          if (!params.project) {
            throw new Error('Project is required!.');
          }
          try {
            const { errors } = await api.deleteClusterMSv({
              serviceName: parseName(showDeleteDialog),
              clusterName: parseName(cluster),
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
      <HandleBackendService
        {...{
          isUpdate: true,
          visible: !!visible,
          setVisible: () => setVisible(null),
          data: visible!,
          templates: templates || [],
        }}
      />
    </>
  );
};

export default BackendServicesResourcesV2;
