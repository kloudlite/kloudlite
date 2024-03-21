import { Link, useParams } from '@remix-run/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
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
import { registryHost } from '~/root/lib/configs/base-url.cjs';
import { handleError } from '~/root/lib/utils/common';
import { Trash } from '~/console/components/icons';
import ListV2 from '~/console/components/listV2';
import ConsoleAvatar from '~/console/components/console-avatar';

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
  const url = `${registryHost}/${account}/${name}`;

  return (
    <CopyContentToClipboard
      content={url}
      toastMessage="Repository url copied successfully."
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
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            render: () => (
              <div className="flex flex-row">
                <span className="w-[48px]" />
                Name
              </div>
            ),
            name: 'name',
            className: 'w-[180px]',
          },
          {
            render: () => 'Repository Url',
            name: 'repositoryUrl',
            className: 'flex-1 w-[180px]',
          },
          {
            render: () => 'Account',
            name: 'account',
            className: 'w-[180px]',
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
                render: () => (
                  <ListTitle
                    title={name}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              repositoryUrl: {
                render: () => (
                  <div className="flex w-fit">
                    <RepoUrlView name={name} />
                  </div>
                ),
              },
              account: { render: () => <ListItem data={i.accountName} /> },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => <ExtraButton onDelete={() => onDelete?.(i)} />,
              },
            },
            to: `/${account}/repo/${btoa(name)}`,
          };
        }),
      }}
    />
  );
};

const RepoResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
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

export default RepoResourcesV2;
