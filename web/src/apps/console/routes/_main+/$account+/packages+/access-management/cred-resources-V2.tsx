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
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ICRCreds } from '~/console/server/gql/queries/cr-queries';
import {
  ExtractNodeType,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { registryHost } from '~/root/lib/configs/base-url.cjs';
import { handleError } from '~/root/lib/utils/common';
import { Copy, Spinner, Trash, Check } from '~/console/components/icons';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import ListV2 from '~/console/components/listV2';
import ConsoleAvatar from '~/console/components/console-avatar';

const RESOURCE_NAME = 'credential';
type BaseType = ExtractNodeType<ICRCreds>;

interface IResource {
  items: BaseType[];
  onDelete: (item: BaseType) => void;
}

const parseAccess = (access: string) => {
  switch (access) {
    case 'read':
      return 'Read Only';
    case 'read_write':
      return 'Read & Write';
    default:
      return 'unknown';
  }
};
const parseItem = (item: BaseType) => {
  return {
    name: item.name,
    id: item.id,
    username: item.username,
    access: parseAccess(item.access),
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

const RegistryUrlView = () => {
  return (
    <CopyContentToClipboard
      content={registryHost}
      toastMessage="Registry url copied successfully."
    />
  );
};

const TokenView = ({ username }: { username: string }) => {
  const iconSize = 16;
  const api = useConsoleApi();
  const [copy, setCopy] = useState(false);
  const [copyIcon, setCopyIcon] = useState(<Copy size={iconSize} />);

  const handleCopy = () => {
    toast.success('Token copied successfully.');
    setCopyIcon(<Check size={iconSize} />);
    setTimeout(() => {
      setCopyIcon(<Copy size={iconSize} />);
    }, 3000);
  };

  const { copy: copyToClipboard } = useClipboard({
    onSuccess: () => {
      handleCopy();
    },
  });

  useDebounce(
    async () => {
      if (copy) {
        setCopyIcon(
          <span className="flex animate-spin">
            <Spinner size={iconSize} />
          </span>
        );
        const { errors, data } = await api.getCredToken({ username });
        if (errors) {
          throw errors[0];
        }
        copyToClipboard(data);
        setCopy(false);
      }
    },
    100,
    [username, copy]
  );
  return (
    <ListBody
      data={
        <div
          className="cursor-pointer flex flex-row items-center gap-lg truncate w-fit hover:text-text-default"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            setCopy(true);
          }}
          title="Copy token"
        >
          <span>Copy token</span>
          {copyIcon}
        </div>
      }
    />
  );
};

const GridView = ({ items, onDelete = (_) => _ }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, access, updateInfo, username } = parseItem(item);
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
                    subtitle={username}
                    action={
                      <div className="flex flex-row items-center">
                        <ExtraButton
                          onDelete={() => {
                            onDelete(item);
                          }}
                        />
                      </div>
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'registry-url'),
                render: () => <RegistryUrlView />,
              },
              {
                key: generateKey(keyPrefix, 'token'),
                render: () => <TokenView username={username} />,
              },
              {
                key: generateKey(keyPrefix, 'access'),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={access} />
                  </div>
                ),
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

const ListView = ({ items, onDelete = (_) => _ }: IResource) => {
  return (
    <ListV2.Root
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
            render: () => 'Registry Url',
            name: 'registryUrl',
            className: 'flex-1 w-[180px]',
          },
          {
            render: () => 'Token',
            name: 'token',
            className: 'w-[180px]',
          },
          {
            render: () => 'Access',
            name: 'access',
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
          const { name, id, access, updateInfo, username } = parseItem(i);
          return {
            columns: {
              name: {
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={username}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              registryUrl: {
                render: () => (
                  <div className="flex w-fit">
                    <RegistryUrlView />
                  </div>
                ),
              },
              token: { render: () => <TokenView username={username} /> },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              access: {
                render: () => <ListBody data={access} />,
              },
              action: {
                render: () => (
                  <ExtraButton
                    onDelete={() => {
                      onDelete(i);
                    }}
                  />
                ),
              },
            },
          };
        }),
      }}
    />
  );
};

const CredResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
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
            const { errors } = await api.deleteCred({
              username: showDeleteDialog!.username,
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

export default CredResourcesV2;
