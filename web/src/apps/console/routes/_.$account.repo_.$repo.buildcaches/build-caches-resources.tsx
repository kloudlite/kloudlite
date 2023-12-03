import { Trash, PencilLine } from '@jengaicons/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IBuildCaches } from '~/console/server/gql/queries/build-caches-queries';
import {
  ExtractNodeType,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import HandleBuildCache from './handle-build-cache';

type BaseType = ExtractNodeType<IBuildCaches>;
const RESOURCE_NAME = 'build cache';

const parseItem = (item: BaseType) => {
  return {
    name: item.name,
    displayName: item.displayName,
    size: item.volumeSizeInGB,
    id: item.id,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IExtraButton {
  onDelete: () => void;
  onEdit: () => void;
}

const ExtraButton = ({ onDelete, onEdit }: IExtraButton) => {
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
}

const GridView = ({ items, onDelete, onEdit }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, displayName, id, updateInfo, size } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={displayName}
                    subtitle={name}
                    action={
                      <ExtraButton
                        onDelete={() => {
                          onDelete(item);
                        }}
                        onEdit={() => onEdit(item)}
                      />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 1),
                render: () => <ListBody data={`${size} GB`} />,
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

const ListView = ({ items, onDelete, onEdit }: IResource) => {
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, displayName, id, updateInfo, size } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, 0),
                className: 'flex-1',
                render: () => <ListTitle subtitle={name} title={displayName} />,
              },
              {
                key: generateKey(keyPrefix, 1),
                className: 'w-[200px]',
                render: () => <ListBody data={`${size} GB`} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[180px]',
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
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
                    onEdit={() => onEdit(item)}
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

const BuildCachesResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleBuildCache, setShowHandleBuildCache] =
    useState<BaseType | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onEdit: (item) => {
      setShowHandleBuildCache(item);
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
            const { errors } = await api.deleteBuildCache({
              crDeleteBuildCacheKeyId: showDeleteDialog!.id,
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
      <HandleBuildCache
        {...{
          isUpdate: true,
          data: showHandleBuildCache!,
          visible: !!showHandleBuildCache,
          setVisible: () => setShowHandleBuildCache(null),
        }}
      />
    </>
  );
};

export default BuildCachesResources;
