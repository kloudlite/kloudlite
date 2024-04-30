import {
  ArrowDown,
  ArrowUp,
  PencilLine,
  Trash,
} from '~/console/components/icons';
import { useEffect, useState } from 'react';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { useReload } from '~/lib/client/helpers/reloader';
import { handleError } from '~/lib/utils/common';
import { IRouter, IRouters } from '~/console/server/gql/queries/router-queries';
import { NN } from '~/lib/types/common';
import { useParams } from '@remix-run/react';
import { toast } from '~/components/molecule/toast';
import HandleRoute from './handle-route';
import { ModifiedRouter } from './_index';

const RESOURCE_NAME = 'domain';
type BaseType = NN<ExtractNodeType<IRouters>['spec']['routes']>[number] & {
  id: string;
};

const parseItem = (item: BaseType) => {
  return {
    path: item.path,
    app: item.app,
    port: item.port,
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'edit' | 'move-down' | 'move-up' | 'delete';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
  isFirst: boolean;
  isLast: boolean;
};

const ExtraButton = ({ onAction, item, isFirst, isLast }: IExtraButton) => {
  let options: IResourceExtraItem[] = [
    {
      label: 'Edit',
      icon: <PencilLine size={16} />,
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
  ];
  if (!isLast) {
    options = [
      {
        label: 'Move down',
        icon: <ArrowDown size={16} />,
        type: 'item',
        onClick: () => onAction({ action: 'move-down', item }),
        key: 'move-down',
      },
      ...options,
    ];
  }
  if (!isFirst) {
    options = [
      {
        label: 'Move up',
        icon: <ArrowUp size={16} />,
        type: 'item',
        onClick: () => onAction({ action: 'move-up', item }),
        key: 'move-up',
      },
      ...options,
    ];
  }
  return <ResourceExtraAction options={options} />;
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items, onAction }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { path, app, port } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${path}-${app}-${index}`;
        return (
          <Grid.Column
            key={path + app + port + item.id}
            rows={[
              {
                key: generateKey(keyPrefix, path),
                render: () => (
                  <ListTitle
                    title={path}
                    action={
                      <ExtraButton
                        onAction={onAction}
                        item={item}
                        isFirst={index === 0}
                        isLast={index === items.length - 1}
                      />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, `${app}`),
                className: 'flex-1',
                render: () => <ListBody data={app} />,
              },
              {
                key: generateKey(keyPrefix, port),
                className: 'flex-1',
                render: () => <ListBody data={port} />,
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
        const { path, app, port } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${path}-${app}-${index}`;
        return (
          <List.Row
            key={path + app + port + item.id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, path),
                className: 'flex-1',
                render: () => <ListTitle title={path} />,
              },
              {
                key: generateKey(keyPrefix, `${app}`),
                className: 'flex-1',
                render: () => <ListBody data={app} />,
              },
              {
                key: generateKey(keyPrefix, port),
                className: 'flex-1',
                render: () => <ListBody data={`:${port}`} />,
              },
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => (
                  <ExtraButton
                    onAction={onAction}
                    item={item}
                    isFirst={index === 0}
                    isLast={index === items.length - 1}
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

const moveItemInArray = (
  arr: BaseType[],
  fromIndex: number,
  toIndex: number
) => {
  const itemToMove = arr.splice(fromIndex, 1)[0]; // Remove the item from the original position
  arr.splice(toIndex, 0, itemToMove); // Insert the item at the new position
  return arr; // Return the modified array
};

const RouteResources = ({
  items = [],
  router,
}: {
  items: BaseType[];
  router?: ModifiedRouter;
}) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [modifiedItems, setModifiedItems] = useState<BaseType[]>([]);
  const [visible, setVisible] = useState<BaseType | null>(null);
  const [moveDir, setMoveDir] = useState('');
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { environment } = useParams();
  useEffect(() => {
    setModifiedItems(items);
  }, [items]);

  const moveItem = (dir: 'move-down' | 'move-up', item: BaseType) => {
    const currIndex = items.findIndex((i) => i.id === item.id);
    switch (dir) {
      case 'move-down':
        if (currIndex < items.length - 1) {
          setModifiedItems([
            ...moveItemInArray(items, currIndex, currIndex + 1),
          ]);
        }
        break;
      case 'move-up':
        if (currIndex > 0) {
          setModifiedItems([
            ...moveItemInArray(items, currIndex, currIndex - 1),
          ]);
        }
        break;
      default:
        break;
    }
  };

  useEffect(() => {
    if (moveDir) {
      setMoveDir('');
      (async () => {
        if ( !environment) {
          throw new Error('Project and Environment is required!');
        }
        if (!router || !router.metadata || !router.spec) {
          throw new Error('Router is required!');
        }
        try {
          const { errors } = await api.updateRouter({
            envName: environment,
            
            router: {
              displayName: router?.displayName,
              spec: {
                ...router?.spec,
                routes: modifiedItems.map((mi) => ({
                  path: mi.path,
                  port: mi.port,
                  app: mi.app,
                })),
              },
              metadata: {
                ...router?.metadata,
              },
            },
          });
          if (errors) {
            throw errors[0];
          }
        } catch (err) {
          handleError(err);
        }
      })();
    }
  }, [modifiedItems, moveDir]);

  const props: IResource = {
    items: modifiedItems,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'edit':
          setVisible(item);
          break;
        case 'delete':
          setShowDeleteDialog(item);
          break;
        case 'move-down':
        case 'move-up':
          moveItem(action, item);
          setMoveDir(action);
          break;
        default:
      }
    },
  };
  return (
    <>
      <ListGridView
        listView={<ListView {...props} items={modifiedItems} />}
        gridView={<GridView {...props} />}
      />
      <DeleteDialog
        resourceName={showDeleteDialog?.path}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            if ( !environment) {
              throw new Error('Project and Environment is required!');
            }
            if (!router || !router.metadata || !router.spec) {
              throw new Error('Router is required!');
            }
            const { errors } = await api.updateRouter({
              envName: environment,
              
              router: {
                displayName: router?.displayName,
                spec: {
                  ...router?.spec,
                  routes: modifiedItems
                    .filter((mi) => mi.id !== showDeleteDialog?.id)
                    .map((mi) => ({
                      path: mi.path,
                      port: mi.port,
                      app: mi.app,
                    })),
                },
                metadata: {
                  ...router?.metadata,
                },
              },
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
      <HandleRoute
        {...{
          isUpdate: true,
          data: visible!,
          visible: !!visible,
          setVisible: () => setVisible(null),
          router,
        }}
      />
    </>
  );
};

export default RouteResources;
