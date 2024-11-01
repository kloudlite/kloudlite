import { Link } from '@remix-run/react';
import { useState } from 'react';
import { toast } from '@kloudlite/design-system/molecule/toast';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItemV2,
  ListTitleV2,
  listClass,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import { Trash } from '~/console/components/icons';
import ListGridView from '~/console/components/list-grid-view';
import ListV2 from '~/console/components/listV2';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IRegistryImages } from '~/console/server/gql/queries/registry-image-queries';
import {
  ExtractNodeType,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';

const RESOURCE_NAME = 'image';
type BaseType = ExtractNodeType<IRegistryImages>;

const parseItem = (item: BaseType) => {
  return {
    name: item.imageName,
    id: item.imageName,
    updateInfo: {
      // author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
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

const ExtraButton = ({ item, onAction }: IExtraButton) => {
  const iconSize = 16;
  const options: IResourceExtraItem[] = [
    {
      label: 'Delete',
      icon: <Trash size={iconSize} />,
      type: 'item',
      onClick: () => onAction({ action: 'delete', item }),
      key: 'delete',
      className: '!text-text-critical',
    },
  ];

  return <ResourceExtraAction options={options} />;
};

interface IResource {
  items: (BaseType & { isClusterOnline: boolean })[];
  onAction: OnAction;
}

const ListView = ({ items, onAction }: IResource) => {
  return (
    <ListV2.Root
      linkComponent={Link}
      data={{
        headers: [
          {
            // render: () => 'Image Name',
            render: () => (
              <div className="flex flex-row">
                <span className="w-[48px]" />
                Image Details
              </div>
            ),
            name: 'name',
            className: listClass.title,
          },
          {
            render: () => 'Registry',
            name: 'registry',
            className: `${listClass.item} flex-1`,
          },
          // {
          //   render: () => '',
          //   name: 'flex-post',
          //   className: listClass.flex,
          // },
          {
            render: () => 'Repository',
            name: 'repository',
            className: `${listClass.item} flex-1`,
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
          const { name, id, updateInfo } = parseItem(i);

          return {
            columns: {
              name: {
                render: () => (
                  <ListTitleV2
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              registry: {
                render: () => <ListItemV2 data={i.meta.registry} />,
              },
              repository: {
                render: () => <ListItemV2 data={i.meta.repository} />,
              },
              updated: {
                render: () => <ListItemV2 subtitle={updateInfo.time} />,
              },
              action: {
                render: () => <ExtraButton item={i} onAction={onAction} />,
              },
            },
            // ...(i.isArchived ? {} : { to: `/${account}/env/${id}` }),
          };
        }),
      }}
    />
  );
};

const ImagesResource = ({ items = [] }: { items: BaseType[] }) => {
  // const { account } = useOutletContext<IAccountContext>();
  const api = useConsoleApi();
  const reloadPage = useReload();

  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const props: IResource = {
    // @ts-ignore
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
        gridView={<ListView {...props} />}
      />
      <DeleteDialog
        resourceName={showDeleteDialog?.imageName || ''}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteRegistryImage({
              image:
                `${showDeleteDialog?.imageName}:${showDeleteDialog?.imageTag}` ||
                '',
            });

            if (errors) {
              throw errors[0];
            }
            reloadPage();
            toast.success(`Image deleted successfully`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};

export default ImagesResource;
