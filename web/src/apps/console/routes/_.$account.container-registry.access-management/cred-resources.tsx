import { CaretDownFill, Copy, Eye, Spinner, Trash } from '@jengaicons/react';
import { useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import ScrollArea from '~/components/atoms/scroll-area';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import {
  ListBody,
  ListItemWithSubtitle,
  ListTitleWithSubtitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
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
import { handleError } from '~/root/lib/utils/common';

const RESOURCE_NAME = 'credential';
type BaseType = ExtractNodeType<ICRCreds>;

interface IResource {
  items: BaseType[];
  onDelete: (item: BaseType) => void;
}

const parseAccess = (access: string) => {
  switch (access) {
    case 'read':
      return 'Read';
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
      author: titleCase(
        `${parseUpdateOrCreatedBy(item)} updated the ${RESOURCE_NAME}`
      ),
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

const TokenView = ({
  username,
  list = true,
}: {
  username: string;
  list: boolean;
}) => {
  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Token copied successfully.');
    },
  });
  const api = useConsoleApi();
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);

  useDebounce(
    async () => {
      if (open) {
        setLoading(true);
        const { errors, data } = await api.getCredToken({ username });
        setLoading(false);
        if (errors) {
          throw errors[0];
        }

        setToken(data);
      }
    },
    100,
    [username, open]
  );
  return (
    <OptionList.Root
      onOpenChange={(e) => {
        setOpen(e);
      }}
    >
      {list ? (
        <OptionList.Trigger>
          <Button
            variant="plain"
            className="group/view"
            content={
              <div className="flex flex-row items-center gap-xl">
                <span>View</span>
                <span className="invisible group-hover/view:visible group-[.selected]/view:visible">
                  <CaretDownFill size={16} />
                </span>
              </div>
            }
            prefix={<Eye />}
          />
        </OptionList.Trigger>
      ) : (
        <OptionList.Trigger>
          <IconButton variant="plain" icon={<Eye />} size="sm" />
        </OptionList.Trigger>
      )}

      <OptionList.Content>
        <OptionList.Item
          onClick={(e) => {
            e.preventDefault();
            copy(token);
          }}
        >
          <ScrollArea className="!w-auto max-w-[200px] min-w-[200px] text-text-default">
            {token}
            {loading && (
              <div className="flex flex-row gap-xl">
                <div className="animate-spin">
                  <Spinner size={16} />
                </div>{' '}
                Loading...
              </div>
            )}
          </ScrollArea>
          <Copy size={16} />
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
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
                  <ListTitleWithSubtitle
                    title={name}
                    subtitle={username}
                    action={
                      <div className="flex flex-row items-center">
                        <TokenView username={item.username} list={false} />
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
                  <ListItemWithSubtitle
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
    <List.Root>
      {items.map((item, index) => {
        const { name, id, access, updateInfo, username } = parseItem(item);
        const keyPrefix = `app-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitle title={name} subtitle={username} />
                ),
              },
              {
                key: generateKey(keyPrefix, 'token'),
                className: 'w-[120px]',
                render: () => <TokenView username={username} list />,
              },
              {
                key: generateKey(keyPrefix, 'access'),
                className: 'w-[120px] text-start',
                render: () => <ListBody data={access} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItemWithSubtitle
                    data={updateInfo.author}
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

const CredResources = ({ items = [] }: { items: BaseType[] }) => {
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
              username: showDeleteDialog?.username || '',
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

export default CredResources;
