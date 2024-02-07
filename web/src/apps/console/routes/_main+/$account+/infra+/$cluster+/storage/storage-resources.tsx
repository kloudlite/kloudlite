import { generateKey } from '~/components/utils';
import {
  ListBody,
  ListSecondary,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IPvs } from '~/console/server/gql/queries/pv-queries';
import { CircleFill, Database, Trash } from '@jengaicons/react';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useState } from 'react';

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
  return (
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
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => <ListTitle title={name} subtitle={id} />,
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

const ListView = ({ items, onAction }: IResource) => {
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id, updateInfo, storage, storageClass, phase } =
          parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'w-[180px] min-w-[180px] max-w-[180px]',
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
              {
                key: generateKey(keyPrefix, 'flex-1'),
                className: 'flex-grow',
                render: () => <div />,
              },
              {
                key: generateKey(keyPrefix, 'phase'),
                className: 'flex-grow',
                render: () => <ListSecondary title="Phase" subtitle={phase} />,
              },
              {
                key: generateKey(keyPrefix, 'flex-2'),
                className: 'flex-grow',
                render: () => <div />,
              },
              {
                key: generateKey(keyPrefix, 'time'),
                className: 'max-w-[180px] w-[180px]',
                render: () => <ListBody data={`Updated ${updateInfo.time}`} />,
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

const StorageResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
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
      {/* <DeleteDialog
        resourceName={parseName(showDeleteDialog)}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          if (!params.project || !params.environment) {
            throw new Error('Project and Environment is required!.');
          }
          try {
            const { errors } = await api.deleteManagedResource({
              mresName: parseName(showDeleteDialog),
              envName: params.environment,
              projectName: params.project,
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
      /> */}
    </>
  );
};

export default StorageResources;
