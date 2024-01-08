import { Copy, Trash } from '@jengaicons/react';
import { Link, useParams } from '@remix-run/react';
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
import { IRepos } from '~/console/server/gql/queries/repo-queries';
import {
  ExtractNodeType,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { registryHost } from '~/root/lib/configs/base-url.cjs';
import { handleError } from '~/root/lib/utils/common';

type BaseType = ExtractNodeType<IRepos>;
const RESOURCE_NAME = 'repository';

const parseItem = (item: BaseType) => {
  return {
    name: item.name,
    id: item.id,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({ onDelete }: { onDelete: () => void }) => {
  return (
    <ResourceExtraAction
      options={[
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

const RepoUrlView = ({ name }: { name: string }) => {
  const { account } = useParams();
  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Registry url copied successfully.');
    },
  });
  const url = `${registryHost}/${account}/${name}`;
  return (
    <ListBody
      data={
        <div
          className="cursor-pointer flex flex-row items-center gap-lg truncate"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            copy(url);
          }}
          title={url}
        >
          <span className="truncate">{url}</span>
          <span>
            <Copy size={16} />
          </span>
        </div>
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
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`../repo/${name}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    action={<ExtraButton onDelete={() => onDelete?.(item)} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'repo-url'),
                render: () => <RepoUrlView name={name} />,
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

const ListView = ({ items, onDelete }: IResource) => {
  const { account } = useParams();
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            to={`/${account}/repo/${name}`}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1 min-w-[100px] max-w-[100px]',
                render: () => <ListTitle title={name} />,
              },
              {
                key: generateKey(keyPrefix, 'repo-url'),
                className: 'min-w-[100px] w-[100px] basis-full  mr-[20px]',
                render: () => <RepoUrlView name={name} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'min-w-[180px] max-w-[180px] text-left',
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => <ExtraButton onDelete={() => onDelete?.(item)} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const RepoResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: setShowDeleteDialog,
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
            const { errors } = await api.deleteRepo({
              name: showDeleteDialog!.name,
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

export default RepoResources;
