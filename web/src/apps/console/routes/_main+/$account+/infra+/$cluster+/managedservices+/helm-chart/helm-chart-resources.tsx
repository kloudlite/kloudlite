import { Trash } from '@jengaicons/react';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
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
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useState } from 'react';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useParams } from '@remix-run/react';
import { IHelmCharts } from '~/console/server/gql/queries/helm-chart-queries';

const RESOURCE_NAME = 'managed service';
type BaseType = ExtractNodeType<IHelmCharts>;

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
    <List.Root>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;

        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name),
                className: 'flex-1',
                render: () => <ListTitle title={name} subtitle={id} />,
              },
              {
                key: generateKey(keyPrefix, 'author'),
                className: 'w-[180px]',
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
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

const HelmChartResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const api = useConsoleApi();
  const reloadPage = useReload();
  const params = useParams();

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
          try {
            const { errors } = await api.deleteClusterMSv({
              serviceName: parseName(showDeleteDialog),
              clusterName: params.cluster || '',
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

export default HelmChartResources;
