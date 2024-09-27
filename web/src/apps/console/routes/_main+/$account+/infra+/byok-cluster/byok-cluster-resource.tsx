import { PencilLine, Trash } from '~/console/components/icons';
import { Link, useParams } from '@remix-run/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import ListV2 from '~/console/components/listV2';
import { IByocClusters } from '~/console/server/gql/queries/byok-cluster-queries';
import AnimateHide from '~/components/atoms/animate-hide';
import LogComp from '~/root/lib/client/components/logger';
import LogAction from '~/console/page-components/log-action';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { useState } from 'react';
import { useDataState } from '~/console/page-components/common-state';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import DeleteDialog from '~/console/components/delete-dialog';
import { useReload } from '~/root/lib/client/helpers/reloader';
import HandleByokCluster from './handle-byok-cluster';

type BaseType = ExtractNodeType<IByocClusters>;
const RESOURCE_NAME = 'byoc clusters';

const parseItem = (item: ExtractNodeType<IByocClusters>) => {
  return {
    name: item.displayName,
    id: parseName(item),
    // path: `/projects/${item.name}`,
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({
  onDelete,
  onEdit,
}: {
  onDelete: () => void;
  onEdit: () => void;
}) => {
  return (
    <ResourceExtraAction
      options={[
        {
          key: '1',
          label: 'Edit',
          icon: <PencilLine size={16} />,
          type: 'item',
          onClick: onEdit,
        },
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

interface IResource {
  items: BaseType[];
  onDelete: (item: BaseType) => void;
  onEdit: (item: BaseType) => void;
}

const GridView = ({ items = [], onEdit, onDelete }: IResource) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            to={`/${account}/${id}/deployment/${id}`}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton
                        onDelete={() => onDelete(item)}
                        onEdit={() => onEdit(item)}
                      />
                    }
                    avatar={<ConsoleAvatar name={id} />}
                  />
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

const ListView = ({ items = [], onEdit, onDelete }: IResource) => {
  const [open, setOpen] = useState<string>('');
  const { state } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');
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
          // {
          //   render: () => '',
          //   name: 'logs',
          //   className: 'w-[180px]',
          // },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
          },
          // {
          //   render: () => 'Provider (Region)',
          //   name: 'provider',
          //   className: 'w-[180px]',
          // },
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
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              status: {
                render: () => <SyncStatusV2 item={i} />,
              },
              // provider: { render: () => <ListItem data={provider} /> },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              action: {
                render: () => (
                  <ExtraButton
                    onDelete={() => onDelete(i)}
                    onEdit={() => onEdit(i)}
                  />
                ),
              },
            },
            detail: (
              <AnimateHide
                onClick={(e) => e.preventDefault()}
                show={open === i.id}
                className="w-full flex pt-4xl pb-2xl justify-center items-center"
              >
                <LogComp
                  {...{
                    hideLineNumber: !state.linesVisible,
                    hideTimestamp: !state.timestampVisible,
                    className: 'flex-1',
                    dark: true,
                    width: '100%',
                    height: '40rem',
                    title: 'Logs',
                    websocket: {
                      account: account || '',
                      cluster: parseName(i),
                      trackingId: i.id,
                    },
                    actionComponent: <LogAction />,
                  }}
                />
              </AnimateHide>
            ),
            hideDetailSeperator: true,
          };
        }),
      }}
    />
  );
};

const ByokClusterResource = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleByokCluster, setShowHandleByokCluster] =
    useState<BaseType | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onEdit: (item) => {
      setShowHandleByokCluster(item);
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
            const { errors } = await api.deleteByokCluster({
              name: parseName(showDeleteDialog),
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
      <HandleByokCluster
        {...{
          isUpdate: true,
          visible: !!showHandleByokCluster,
          setVisible: () => {
            setShowHandleByokCluster(null);
          },
          data: showHandleByokCluster!,
        }}
      />
    </>
  );
};

export default ByokClusterResource;
