import { Trash } from '@jengaicons/react';
import { useOutletContext, useParams } from '@remix-run/react';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { generateKey, titleCase } from '~/components/utils';
import List from '~/console/components/list';
import {
  ExtractNodeType,
  parseName,
  parseTargetNs,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '../components/console-list-components';
import DeleteDialog from '../components/delete-dialog';
import Grid from '../components/grid';
import ListGridView from '../components/list-grid-view';
import ResourceExtraAction from '../components/resource-extra-action';
import { useConsoleApi } from '../server/gql/api-provider';
import { ISecrets } from '../server/gql/queries/secret-queries';
import { IWorkspaceContext } from '../routes/_main+/$account+/$cluster+/$project+/$scope+/$workspace+/_layout';

const RESOURCE_NAME = 'secret';
type BaseType = ExtractNodeType<ISecrets>;

interface IResource {
  onDelete: (item: BaseType) => void;
  hasActions?: boolean;
  onClick?: (item: BaseType) => void;
  linkComponent?: any;
  items: BaseType[];
}

const parseItem = (item: BaseType) => {
  return {
    name: titleCase(item.displayName),
    id: parseName(item),
    entries: [`${Object.keys(item?.stringData || {}).length || 0} Entries`],
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
  const { account, cluster, project, scope, workspace } = useParams();
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
                ? `/${account}/${cluster}/${project}/${scope}/${workspace}/secret/${id}`
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
  const { account, cluster, project, scope, workspace } = useParams();
  const [selected, setSelected] = useState('');
  let props = {};
  if (linkComponent) {
    props = { linkComponent };
  }
  return (
    <List.Root {...props}>
      {items.map((item, index) => {
        const { name, id, entries, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;

        return (
          <List.Row
            onClick={() => {
              onClick(item);
              setSelected(id);
            }}
            pressed={selected === id}
            key={id}
            className="!p-3xl"
            to={
              linkComponent !== null
                ? `/${account}/${cluster}/${project}/${scope}/${workspace}/secret/${id}`
                : undefined
            }
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => <ListTitle title={name} />,
              },
              {
                key: generateKey(keyPrefix, 'entries'),
                className: 'w-[120px]',
                render: () => <ListBody data={entries} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[180px]',
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              ...[
                ...(hasActions
                  ? [
                      {
                        key: generateKey(keyPrefix, 'action'),
                        render: () => (
                          <ExtraButton onDelete={() => onDelete(item)} />
                        ),
                      },
                    ]
                  : []),
              ],
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const SecretResources = ({
  items = [],
  hasActions = true,
  onClick = (_) => _,
  linkComponent = null,
}: Omit<IResource, 'onDelete'>) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const api = useConsoleApi();
  const reloadPage = useReload();
  const { workspace } = useOutletContext<IWorkspaceContext>();

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
          try {
            const { errors } = await api.deleteSecret({
              name: parseName(showDeleteDialog),
              namespace: parseTargetNs(workspace),
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

export default SecretResources;
