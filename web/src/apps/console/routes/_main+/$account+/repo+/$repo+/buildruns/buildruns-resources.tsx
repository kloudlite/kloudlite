import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
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
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { useParams } from '@remix-run/react';
import { IBuildRuns } from '~/console/server/gql/queries/build-run-queries';

const RESOURCE_NAME = 'build run';
type BaseType = ExtractNodeType<IBuildRuns>;

const parseItem = (item: BaseType) => {
  return {
    name: parseName(item),
    id: parseName(item),
    updateInfo: {
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

interface IExtraButton {
  onDelete: () => void;
}
const ExtraButton = ({ onDelete }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={
        [
          // {
          //   label: 'Delete',
          //   icon: <Trash size={16} />,
          //   type: 'item',
          //   onClick: onDelete,
          //   key: 'delete',
          //   className: '!text-text-critical',
          // },
        ]
      }
    />
  );
};

interface IResource {
  items: BaseType[];
  onDelete: (item: BaseType) => void;
}

const GridView = ({ items, onDelete }: IResource) => {
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
                  // action={
                  //   <ExtraButton
                  //     onDelete={() => {
                  //       onDelete(item);
                  //     }}
                  //   />
                  // }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'time'),
                render: () => (
                  <ListBody data={`Last Updated ${updateInfo.time}`} />
                ),
              },
            ]}
          />
        );
      })}
    </Grid.Root>
  );
};

const ListItem = ({ item }: { item: BaseType }) => {
  const { name, id, updateInfo } = parseItem(item);
  return (
    <div className="flex flex-col">
      <div className="flex flex-row">
        <ListTitle title={name} subtitle={id} />
      </div>
      <div className="">Logs</div>
    </div>
  );
};
const ListView = ({ items, onDelete }: IResource) => {
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
                key: generateKey(keyPrefix, name + id),
                render: () => <ListItem item={item} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const BuildRunResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
  };

  const params = useParams();
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
            const { errors } = await api.deleteVpnDevice({
              deviceName: parseName(showDeleteDialog),
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

export default BuildRunResources;
