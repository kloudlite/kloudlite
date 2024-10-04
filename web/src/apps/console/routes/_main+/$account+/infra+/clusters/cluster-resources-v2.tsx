import { Link, useOutletContext, useParams } from '@remix-run/react';
import { generateKey, titleCase, useMapper } from '~/components/utils';
import { listRender } from '~/console/components/commons';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListBody,
  ListItem,
  ListItemV2,
  ListTitle,
  ListTitleV2,
  listClass,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import {
  GearSix,
  ListDashes,
  PencilLine,
  Trash,
} from '~/console/components/icons';
import ListGridView from '~/console/components/list-grid-view';
import ListV2 from '~/console/components/listV2';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { renderCloudProvider } from '~/console/utils/commons';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import logger from '~/root/lib/client/helpers/log';

import { useState } from 'react';
// import { SyncStatusV2 } from '~/console/components/sync-status';
import { Badge } from '~/components/atoms/badge';
import { Button } from '~/components/atoms/button';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import CodeView from '~/console/components/code-view';
import DeleteDialog from '~/console/components/delete-dialog';
import { LoadingPlaceHolder } from '~/console/components/loading';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IByocClusters } from '~/console/server/gql/queries/byok-cluster-queries';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { handleError } from '~/root/lib/utils/common';
// import { Github__Com___Kloudlite___Api___Pkg___Types__SyncState as SyncStatusState } from '~/root/src/generated/gql/server';
import TooltipV2 from '~/components/atoms/tooltipV2';
import { ViewClusterLogs } from '~/console/components/cluster-logs-popop';
import { useClusterStatusV2 } from '~/console/hooks/use-cluster-status-v2';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import HandleByokCluster from '../byok-cluster/handle-byok-cluster';

type BaseType = ExtractNodeType<IClusters> & { type: 'normal' };
type ByokBaseType = ExtractNodeType<IByocClusters> & { type: 'byok' };
type CombinedBaseType = BaseType | ByokBaseType;

const RESOURCE_NAME = 'cluster';

const getProvider = (item: BaseType) => {
  if (!item.spec) {
    return '';
  }
  switch (item.spec.cloudProvider) {
    case 'aws':
      return (
        <span>
          {renderCloudProvider({ cloudprovider: item.spec.cloudProvider })}{' '}
          <span>({item.spec.aws?.region})</span>
        </span>
      );
    case 'gcp':
      return (
        <span>
          {renderCloudProvider({ cloudprovider: item.spec.cloudProvider })}{' '}
          <span>({item.spec.gcp?.region})</span>
        </span>
      );
    case 'azure':
      return <span>{item.spec.cloudProvider}</span>;

    default:
      logger.error('unknown provider', item.spec.cloudProvider);
      return '';
  }
};

const parseItem = (item: CombinedBaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    provider: item.type === 'byok' ? null : getProvider(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ByokInstructionsPopup = ({
  show,
  item,
  onClose,
}: {
  show: boolean;
  item: CombinedBaseType;
  onClose: () => void;
}) => {
  const params = useParams();
  ensureAccountClientSide(params);
  const api = useConsoleApi();

  const { data, isLoading, error } = useCustomSwr(
    item.metadata?.name || null,
    async () => {
      if (!item.metadata?.name) {
        throw new Error('Invalid cluster name');
      }
      return api.getBYOKClusterInstructions({
        name: item.metadata.name,
      });
    }
  );

  return (
    <Popup.Root onOpenChange={onClose} show={show} className="!w-[800px]">
      <Popup.Header>Instructions to attach cluster</Popup.Header>
      <Popup.Content>
        <form className="flex flex-col gap-2xl">
          {error && (
            <span className="bodyMd-medium text-text-strong">
              Error while fetching instructions
            </span>
          )}
          {isLoading ? (
            <LoadingPlaceHolder />
          ) : (
            data && (
              <div className="flex flex-col gap-sm text-start ">
                <span className="flex flex-wrap items-center gap-md py-lg">
                  Please follow below instruction for further steps
                </span>

                {data.map((d, index) => {
                  return (
                    <div key={d.title} className="flex flex-col gap-lg pb-2xl">
                      <span className="bodyMd-medium text-text-strong font-bold">
                        Step {`${index + 1}: ${d.title}`}
                      </span>
                      <CodeView
                        preClassName="!overflow-none text-wrap break-words"
                        copy
                        data={d.command || ''}
                      />
                    </div>
                  );
                })}
              </div>
            )
          )}
        </form>
      </Popup.Content>
      <Popup.Footer>
        <Button variant="primary-outline" content="close" onClick={onClose} />
      </Popup.Footer>
    </Popup.Root>
  );
};

const ByokButton = ({ item }: { item: CombinedBaseType }) => {
  const [show, setShow] = useState(false);

  return (
    <div>
      {show ? (
        <ByokInstructionsPopup
          show={show}
          onClose={() => {
            setShow(false);
          }}
          item={item}
        />
      ) : (
        <div className="flex gap-xl items-center pulsable">
          {/* <span>{item.aws?.awsAccountId}</span> */}
          <Button
            content="setup"
            onClick={() => {
              setShow(true);
            }}
            size="sm"
            variant="outline"
          // prefix={<ArrowClockwise size={16} />}
          />
        </div>
      )}
    </div>
  );
};

const GetByokClusterMessage = ({
  lastOnlineAt,
  item,
}: {
  lastOnlineAt: string;
  item: CombinedBaseType;
}) => {
  if (lastOnlineAt === null) {
    return <ByokButton item={item} />;
  }

  const lastTime = new Date(lastOnlineAt);
  const currentTime = new Date();

  const timeDifference =
    (currentTime.getTime() - lastTime.getTime()) / (1000 * 60);

  switch (true) {
    case timeDifference <= 1:
      return (
        <div className="flex flex-row gap-sm bodyMd-medium text-text-strong pulsable">
          <span>Attached Compute</span>
        </div>
      );
    default:
      return <ByokButton item={item} />;
  }
};

const GetSyncStatus = ({ lastOnlineAt }: { lastOnlineAt: string }) => {
  const tooltipOffset = 5;
  if (lastOnlineAt === null || typeof lastOnlineAt === 'object') {
    return <Badge type="warning">Offline</Badge>;
  }

  const lastTime = new Date(lastOnlineAt);
  const currentTime = new Date();

  const timeDifference =
    (currentTime.getTime() - lastTime.getTime()) / (1000 * 60);

  switch (true) {
    case timeDifference <= 1:
      return (
        <TooltipV2
          className="!w-fit !max-w-[500px]"
          place="top"
          offset={tooltipOffset}
          content={
            <div className="flex-1 bodySm text-text-strong pulsable whitespace-normal">
              Last seen ({Math.floor(timeDifference * 60)}s ago)
            </div>
          }
        >
          <div>
            <Badge type="info">Online</Badge>
          </div>
        </TooltipV2>
      );
    case timeDifference > 1:
      return (
        <TooltipV2
          className="!w-fit !max-w-[500px]"
          place="top"
          offset={tooltipOffset}
          content={
            <div className="flex-1 bodySm text-text-strong pulsable whitespace-normal">
              Last seen ({Math.floor(timeDifference * 60)}s ago)
            </div>
          }
        >
          <div>
            <Badge type="warning">Offline</Badge>
          </div>
        </TooltipV2>
      );
    default:
      return (
        <div>
          <Badge type="warning">Offline</Badge>
        </div>
      );
  }
};

const ExtraButton = ({
  onDelete,
  onEdit,
  onShowLogs,
  item,
}: {
  onDelete: () => void;
  onEdit: () => void;
  onShowLogs: () => void;
  item: CombinedBaseType;
}) => {
  const { account } = useParams();
  return item.type === 'byok' ? (
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
  ) : (
    <ResourceExtraAction
      options={[
        {
          label: 'Show Logs',
          icon: <ListDashes size={16} />,
          type: 'item',
          onClick: onShowLogs,
          key: '1',
          // className: '!text-text-critical',
        },
        {
          label: 'Settings',
          icon: <GearSix size={16} />,
          type: 'item',
          to: `/${account}/infra/${item.metadata.name}/settings`,
          key: 'settings',
        },
      ]}
    />
  );
};

interface IResource {
  items: CombinedBaseType[];
  onDelete: (item: CombinedBaseType) => void;
  onEdit: (item: CombinedBaseType) => void;
  onShowLogs: (item: CombinedBaseType) => void;
}

const GridView = ({ items = [], onEdit, onDelete, onShowLogs }: IResource) => {
  const { account } = useParams();
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3" linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, provider, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        const lR = listRender({ keyPrefix, resource: item });
        const status = lR.statusRender({ className: '' });
        return (
          <Grid.Column
            key={id}
            to={`/${account}/infra/${id}/overview`}
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
                        onShowLogs={() => onShowLogs(item)}
                        item={item}
                      />
                    }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, id + name + provider),
                render: () => (
                  <div className="flex flex-col gap-md">
                    <ListBody data={provider} />
                  </div>
                ),
              },
              status,
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
const ListView = ({ items = [], onEdit, onDelete, onShowLogs }: IResource) => {
  const { account } = useParams();
  const { clusters } = useClusterStatusV2();
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
            className: listClass.title,
          },
          {
            render: () => '',
            name: 'provider',
            className: listClass.flex,
          },
          {
            render: () => 'Status',
            name: 'status',
            className: listClass.status,
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
          const { name, id, updateInfo, provider } = parseItem(i);

          // const isLatest = dayjs(i.updateTime).isAfter(
          //   dayjs().subtract(3, 'hour')
          // );

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
              provider: {
                render: () => {
                  if (i.type === 'byok' && i.ownedBy !== null) {
                    return (
                      <div className="flex flex-row gap-sm bodyMd-medium text-text-strong pulsable">
                        <span>Local Device</span>
                      </div>
                    );
                  }
                  return (
                    <GetByokClusterMessage
                      // lastOnlineAt={i.lastOnlineAt}
                      lastOnlineAt={clusters[id]?.lastOnlineAt}
                      item={i}
                    />
                  );
                },
              },
              status: {
                render: () => (
                  <GetSyncStatus lastOnlineAt={clusters[id]?.lastOnlineAt} />
                ),
              },
              updated: {
                render: () => (
                  <ListItemV2
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
                    onShowLogs={() => onShowLogs(i)}
                    item={i}
                  />
                ),
              },
            },
            ...(i.type === 'normal'
              ? { to: `/${account}/infra/${id}/overview` }
              : {}),
          };
        }),
      }}
    />
  );
};

const ClusterResourcesV2 = ({
  byokItems,
}: {
  byokItems: Omit<ByokBaseType, 'type'>[];
}) => {
  const { account } = useOutletContext<IAccountContext>();

  const bItems = useMapper(byokItems, (i) => {
    return { ...i, type: 'byok' as ByokBaseType['type'] };
  });

  useWatchReload(
    bItems.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(i)}`;
    })
  );

  const [showDeleteDialog, setShowDeleteDialog] =
    useState<CombinedBaseType | null>(null);
  const [showHandleByokCluster, setShowHandleByokCluster] =
    useState<CombinedBaseType | null>(null);
  const [showClusterLogs, setShowClusterLogs] =
    useState<CombinedBaseType | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items: bItems,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onEdit: (item) => {
      setShowHandleByokCluster(item);
    },
    onShowLogs: (item) => {
      setShowClusterLogs(item);
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
      {showClusterLogs && (
        <ViewClusterLogs
          show={!!showClusterLogs}
          setShow={() => {
            setShowClusterLogs(null);
          }}
          item={showClusterLogs!}
        />
      )}
      <HandleByokCluster
        {...{
          isUpdate: true,
          visible: !!showHandleByokCluster,
          setVisible: () => {
            setShowHandleByokCluster(null);
          },
          data: showHandleByokCluster! as ByokBaseType,
        }}
      />
    </>
  );
};

export default ClusterResourcesV2;
