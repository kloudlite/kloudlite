import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IPvs } from '~/console/server/gql/queries/pv-queries';
import { CircleFill, Database, Trash } from '~/console/components/icons';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useState } from 'react';
import DeleteDialog from '~/console/components/delete-dialog';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useOutletContext, useParams } from '@remix-run/react';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';

const RESOURCE_NAME = 'storage';
type BaseType = ExtractNodeType<IPvs>;

const parseItem = (item: BaseType) => {
  return {
    name: parseName(item),
    id: parseName(item),
    storage: item.spec?.capacity?.storage,
    storageClass: item.spec?.storageClassName,
    phase: item.status?.phase,
    updateInfo: {
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'delete';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
};

const ExtraButton = ({ onAction, item }: IExtraButton) => {
  const phase = item.status?.phase;
  return phase !== 'Bound' ? (
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
  ) : null;
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items, onAction }: IResource) => {
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
                    subtitle={id}
                    action={<ExtraButton item={item} onAction={onAction} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'time'),
                render: () => <ListBody data={`Updated ${updateInfo.time}`} />,
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
            render: () => (
              <div className="flex flex-row">
                <span className="w-[32px]" />
                Name
              </div>
            ),
            name: 'name',
            className: 'w-[300px]',
          },
          {
            render: () => 'Phase',
            name: 'phase',
            className: 'w-[180px] ml-[20px]',
          },
          {
            render: () => '',
            name: 'flex',
            className: 'flex-1',
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
          const { name, updateInfo, storage, storageClass, phase } =
            parseItem(i);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={
                      <div className="flex flex-row items-baseline gap-sm">
                        {storage}
                        <CircleFill size={7} />
                        <span className="bodySm-medium text-text-strong">
                          {storageClass}
                        </span>
                      </div>
                    }
                    avatar={
                      <span className="pulsable pulsable-img min-h-3xl">
                        <Database size={20} />
                      </span>
                    }
                  />
                ),
              },
              phase: {
                render: () => <ListItem data={phase} />,
              },
              flex: {
                render: () => null,
              },
              updated: {
                render: () => (
                  <ListItem subtitle={`Updated ${updateInfo.time}`} />
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
const StorageResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const { cluster } = useParams();
  const reloadPage = useReload();
  const api = useConsoleApi();

  const { account } = useOutletContext<IAccountContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(
        account
      )}.cluster:${cluster}.persistance_volume:${parseName(i)}`;
    })
  );

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
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
          if (!cluster) {
            throw new Error('Cluster is required!.');
          }
          try {
            const { errors } = await api.deletePV({
              pvName: parseName(showDeleteDialog),
              clusterName: cluster,
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
    </>
  );
};

export default StorageResourcesV2;
