import { Trash, PencilLine } from '@jengaicons/react';
import { useState } from 'react';
import { generateKey } from '~/components/utils';
import {
  ListBody,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { IRouter, IRouters } from '~/console/server/gql/queries/router-queries';
import { NN } from '~/root/lib/types/common';
import HandleRoute from './handle-route';

const RESOURCE_NAME = 'domain';
type BaseType = NN<ExtractNodeType<IRouters>['spec']['routes']>[number];

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
  action: 'edit' | 'delete' | 'detail';
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
        const { path, app, port } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${path}-${app}-${index}`;
        return (
          <Grid.Column
            key={path + app + port}
            rows={[
              {
                key: generateKey(keyPrefix, path),
                render: () => (
                  <ListTitle
                    title={path}
                    action={<ExtraButton onAction={onAction} item={item} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, port),
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
            key={path + app + port}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, path),
                className: 'flex-1',
                render: () => <ListTitle title={path} />,
              },
              {
                key: generateKey(keyPrefix, port),
                className: 'flex-1',
                render: () => <ListBody data={app} />,
              },
              {
                key: generateKey(keyPrefix, port),
                className: 'flex-1',
                render: () => <ListBody data={port} />,
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

const RouteResources = ({
  items = [],
  router,
}: {
  items: BaseType[];
  router: IRouter;
}) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [visible, setVisible] = useState<BaseType | null>(null);
  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'edit':
          setVisible(item);
          break;
        case 'delete':
          setShowDeleteDialog(item);
          break;
        default:
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
        resourceName={showDeleteDialog?.path}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            // const { errors } = await api.deleteDomain({
            //   domainName: showDeleteDialog!.domainName,
            // });

            // if (errors) {
            //   throw errors[0];
            // }
            // reloadPage();
            // toast.success(`${titleCase(RESOURCE_NAME)} deleted successfully`);
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
