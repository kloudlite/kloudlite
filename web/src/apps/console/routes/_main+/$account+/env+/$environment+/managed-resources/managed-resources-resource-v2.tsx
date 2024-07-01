import { LockSimple, Trash } from '~/console/components/icons';
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
import DeleteDialog from '~/console/components/delete-dialog';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/lib/client/helpers/reloader';
import { useState } from 'react';
import { handleError } from '~/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
// import { SyncStatusV2 } from '~/console/components/sync-status';
import { IManagedResources } from '~/console/server/gql/queries/managed-resources-queries';
// import ConsoleAvatar from '~/console/components/console-avatar';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import { getManagedTemplate } from '~/console/utils/commons';
import { ViewSecret } from './handle-managed-resource-v2';

const RESOURCE_NAME = 'managed resource';
type BaseType = ExtractNodeType<IManagedResources>;

const parseItem = (item: BaseType, templates: IMSvTemplates) => {
  const template = getManagedTemplate({
    templates,
    kind: item.spec?.resourceTemplate.msvcRef?.kind || '',
    apiVersion: item.spec?.resourceTemplate.msvcRef?.apiVersion || '',
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
  action: 'delete' | 'edit' | 'view_secret';
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
        // {
        //   label: 'Edit',
        //   icon: <PencilSimple size={16} />,
        //   type: 'item',
        //   onClick: () => onAction({ action: 'edit', item }),
        //   key: 'edit',
        // },
        {
          label: 'View Secret',
          icon: <LockSimple size={16} />,
          type: 'item',
          onClick: () => onAction({ action: 'view_secret', item }),
          key: 'view_secret',
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
  templates: IMSvTemplates;
}

const GridView = ({ items = [], onAction, templates }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item, templates);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={<ExtraButton onAction={onAction} item={item} />}
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

const ListView = ({ items = [], onAction, templates }: IResource) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => 'Resource Name',
            name: 'name',
            className: 'flex flex-1 w-[80px]',
          },
          // {
          //   render: () => '',
          //   name: 'secret',
          //   className: 'flex flex-1 w-[150px]',
          // },
          {
            render: () => 'Resource Type',
            name: 'resource',
            className: 'w-[140px]',
          },
          {
            render: () => '',
            name: 'flex-pre',
            className: 'flex-1',
          },
          {
            render: () => 'Integrated Service',
            name: 'service',
            className: 'w-[200px]',
          },
          {
            render: () => '',
            name: 'flex-post',
            className: 'flex-1',
          },
          // {
          //   render: () => 'Status',
          //   name: 'status',
          //   className: 'flex-1 min-w-[30px]',
          // },
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
                    // avatar={
                    //   <div className="pulsable pulsable-circle aspect-square">
                    //     <img src={i.} alt={name} className="w-4xl h-4xl" />
                    //   </div>
                    // }
                    // avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              secret: {
                render: () =>
                  i.syncedOutputSecretRef ? (
                    <Button
                      content="View secrets"
                      variant="plain"
                      onClick={() =>
                        onAction({ action: 'view_secret', item: i })
                      }
                    />
                  ) : null,
              },
              resource: {
                render: () => (
                  <ListItem data={`${i.spec?.resourceTemplate?.kind}`} />
                ),
              },
              service: {
                render: () => (
                  <ListItem
                    data={
                      <div className="flex flex-row gap-xl">
                        <span>
                          <img
                            src={logo}
                            alt={`${i.spec?.resourceTemplate?.msvcRef?.name}`}
                            className="w-4xl h-4xl"
                          />
                        </span>
                        <span>{`${i.spec?.resourceTemplate?.msvcRef?.name}`}</span>
                      </div>
                    }
                  />
                ),
              },
              // status: {
              //   render: () => <SyncStatusV2 item={i} />,
              // },
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
          };
        }),
      }}
    />
  );
};

const ManagedResourceResourcesV2 = ({
  items = [],
  templates = [],
}: {
  items: BaseType[];
  templates: IMSvTemplates;
}) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showSecret, setShowSecret] = useState<BaseType | null>(null);
  const api = useConsoleApi();
  const reloadPage = useReload();
  const params = useParams();

  const { environment, account } = useParams();

  useWatchReload(
    items.map((i) => {
      return `account:${account}.environment:${environment}.managed_resource:${parseName(
        i
      )}`;
    })
  );

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'delete':
          setShowDeleteDialog(item);
          break;
        case 'view_secret':
          setShowSecret(item);
          break;
        default:
          break;
      }
    },
    templates,
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
          if (!params.environment) {
            throw new Error('Environment is required!.');
          }
          try {
            const { errors } = await api.deleteImportedManagedResource({
              mresName: parseName(showDeleteDialog),
              envName: environment || '',
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

      {showSecret && (
        <ViewSecret
          show={!!showSecret}
          setShow={() => {
            setShowSecret(null);
          }}
          item={showSecret!}
        />
      )}
    </>
  );
};

export default ManagedResourceResourcesV2;
