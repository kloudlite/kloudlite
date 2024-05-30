import { PencilSimple, Trash } from '~/console/components/icons';
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
import DeleteDialog from '~/console/components/delete-dialog';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/lib/client/helpers/reloader';
import { useState } from 'react';
import { handleError } from '~/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useParams } from '@remix-run/react';
import { IManagedResources } from '~/console/server/gql/queries/managed-resources-queries';
import { Button } from '~/components/atoms/button';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
import { SyncStatusV2 } from '~/console/components/sync-status';
import HandleManagedResources, { ViewSecret } from './handle-managed-resource';

const RESOURCE_NAME = 'managed resource';
type BaseType = ExtractNodeType<IManagedResources>;

const parseItem = (item: BaseType) => {
  return {
    name: item?.displayName,
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
  onAction: OnAction;
}

const GridView = ({ items = [], onAction }: IResource) => {
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

const ListView = ({ items = [], onAction }: IResource) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => 'Name',
            name: 'name',
            className: 'w-[80px]',
          },
          {
            render: () => '',
            name: 'secret',
            className: 'flex flex-1 w-[150px]',
          },
          {
            render: () => 'Type',
            name: 'type',
            className: 'w-[100px]',
          },
          {
            render: () => '',
            name: 'flex-pre',
            className: 'flex-1',
          },
          {
            render: () => 'Kind',
            name: 'kind',
            className: 'w-[180px]',
          },
          {
            render: () => '',
            name: 'flex-post',
            className: 'flex-1',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px]',
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
                render: () => <ListTitle title={name} subtitle={id} />,
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
              type: {
                render: () => (
                  <ListItem data={`${i.spec?.resourceTemplate?.kind}`} />
                ),
              },
              kind: {
                render: () => (
                  <ListItem
                    data={`${i.spec?.resourceTemplate?.msvcRef?.kind}`}
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
  const [visible, setVisible] = useState<BaseType | null>(null);
  const api = useConsoleApi();
  const reloadPage = useReload();

  // TODO: nxtcoder36: remvove env once api is updated
  // const { environment, account } = useParams();
  const { account } = useParams();
  const environment = 'need to remove it';

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
        case 'edit':
          setVisible(item);
          break;
        case 'view_secret':
          setShowSecret(item);
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
          // if (!params.project || !params.environment) {
          //   throw new Error('Project and Environment is required!.');
          // }
          try {
            const { errors } = await api.deleteManagedResource({
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
      <HandleManagedResources
        {...{
          isUpdate: true,
          visible: !!visible,
          setVisible: () => setVisible(null),
          data: visible!,
          templates: templates || [],
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
