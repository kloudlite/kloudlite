import { Trash } from '~/console/components/icons';
import { useParams } from '@remix-run/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import {
  ListBody,
  ListItem,
  ListItemV2,
  ListTitle,
  ListTitleV2,
  listClass,
} from '../components/console-list-components';
import DeleteDialog from '../components/delete-dialog';
import Grid from '../components/grid';
import ListGridView from '../components/list-grid-view';
import ResourceExtraAction from '../components/resource-extra-action';
import { useConsoleApi } from '../server/gql/api-provider';
import { IConfigs } from '../server/gql/queries/config-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '../server/r-utils/common';
import ListV2 from '../components/listV2';

const RESOURCE_NAME = 'config';
type BaseType = ExtractNodeType<IConfigs>;

interface IResource {
  onDelete: (item: BaseType) => void;
  hasActions?: boolean;
  onClick?: (item: BaseType) => void;
  linkComponent?: any;
  items: BaseType[];
}

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    entries: `${Object.keys(item?.data).length || 0} Entries`,
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

const GridView = ({
  items = [],
  hasActions = true,
  onClick = (_) => _,
  onDelete = (_) => _,
  linkComponent = null,
}: IResource) => {
  const { account, environment } = useParams();
  const [selected, setSelected] = useState('');
  let props = {};
  if (linkComponent) {
    props = { linkComponent };
  }
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" {...props}>
      {items.map((item, index) => {
        const { name, id, entries, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            onClick={() => {
              onClick(item);
              setSelected(id);
            }}
            pressed={selected === id}
            key={id}
            to={
              linkComponent !== null
                ? `/${account}/env/${environment}/config/${id}`
                : undefined
            }
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    action={
                      hasActions && (
                        <ExtraButton
                          onDelete={() => {
                            onDelete(item);
                          }}
                        />
                      )
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'entries'),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={entries} />
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

const ListView = ({
  items = [],
  hasActions = true,
  onClick = (_) => _,
  onDelete = (_) => _,
  linkComponent = null,
}: IResource) => {
  const { account, environment } = useParams();
  const [selected, setSelected] = useState('');
  let props = {};
  if (linkComponent) {
    props = { linkComponent };
  }
  return (
    <ListV2.Root
      {...props}
      data={{
        headers: [
          {
            render: () => 'Name',
            name: 'name',
            className: listClass.title,
          },
          {
            render: () => 'Entries',
            name: 'entries',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
          },
          {
            render: () => 'Updated',
            name: 'updated',
            className: listClass.updated,
          },
          {
            render: () => '',
            name: 'action',
            className: listClass.action,
          },
        ],
        rows: items.map((i) => {
          const { name, id, entries, updateInfo } = parseItem(i);
          return {
            onClick: () => {
              onClick(i);
              setSelected(id);
            },
            pressed: !linkComponent ? selected === id : false,
            columns: {
              name: {
                render: () => <ListTitleV2 title={name} />,
              },
              entries: {
                render: () => <ListItemV2 data={entries} />,
              },
              updated: {
                render: () => (
                  <ListItemV2
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },

              ...(hasActions
                ? {
                    action: {
                      render: () => (
                        <ExtraButton onDelete={() => onDelete(i)} />
                      ),
                    },
                  }
                : {}),
            },
            to:
              linkComponent !== null
                ? `/${account}/env/${environment}/config/${id}`
                : undefined,
          };
        }),
      }}
    />
  );
};

const ConfigResourcesV2 = ({
  items = [],
  hasActions = true,
  onClick = (_) => _,
  linkComponent = null,
}: Omit<IResource, 'onDelete'>) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null,
  );

  const api = useConsoleApi();
  const reloadPage = useReload();
  const { environment, account } = useParams();

  useWatchReload(
    items.map((i) => {
      return `account:${account}.environment:${environment}.config:${parseName(
        i,
      )}`;
    }),
  );

  const props: IResource = {
    items,
    hasActions,
    onClick,
    linkComponent,
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
        resourceName={showDeleteDialog?.displayName}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          if (!environment) {
            throw new Error('Project and Environment name is required!.');
          }
          try {
            const { errors } = await api.deleteConfig({
              configName: parseName(showDeleteDialog),

              envName: environment,
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

export default ConfigResourcesV2;
