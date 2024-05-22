import {
  ArrowClockwise,
  GearSix,
  PencilLine,
  Trash,
} from '~/console/components/icons';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import {
  generateKey,
  titleCase,
  useAppend,
  useMapper,
} from '~/components/utils';
import { listRender } from '~/console/components/commons';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { IClusters } from '~/console/server/gql/queries/cluster-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { renderCloudProvider } from '~/console/utils/commons';
import logger from '~/root/lib/client/helpers/log';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import ListV2 from '~/console/components/listV2';
import AnimateHide from '~/components/atoms/animate-hide';
import LogComp from '~/root/lib/client/components/logger';
import LogAction from '~/console/page-components/log-action';
import { useDataState } from '~/console/page-components/common-state';
import { useState } from 'react';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { Button } from '~/components/atoms/button';
import { dayjs } from '~/components/molecule/dayjs';
import { IByocClusters } from '~/console/server/gql/queries/byok-cluster-queries';
import DeleteDialog from '~/console/components/delete-dialog';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import Popup from '~/components/molecule/popup';
import CodeView from '~/console/components/code-view';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
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
        <div className="flex flex-row items-center gap-lg">
          {renderCloudProvider({ cloudprovider: item.spec.cloudProvider })}
          <span>({item.spec.aws?.region})</span>
        </div>
      );
    case 'gcp':
    case 'azure':
      return (
        <div className="flex flex-row items-center gap-lg">
          <span>{item.spec.cloudProvider}</span>
        </div>
      );

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
  clusterName,
}: {
  show: boolean;
  item: CombinedBaseType;
  onClose: () => void;
  clusterName: string;
}) => {
  const api = useConsoleApi();

  const { data } = useCustomSwr(item.metadata?.name || null, async () => {
    if (!item.metadata?.name) {
      throw new Error('Invalid cluster name');
    }
    return api.getBYOKClusterInstructions({
      name: item.metadata.name,
    });
  });

  console.log('data', data);

  // const [isLoading, setIsLoading] = useState(false);

  return (
    <Popup.Root onOpenChange={onClose} show={show} className="!w-[800px]">
      <Popup.Header>{`${clusterName} setup instructions:`}</Popup.Header>
      <Popup.Content>
        <form className="flex flex-col gap-2xl">
          {data && (
            <div className="flex flex-col gap-xl text-start ">
              <span className="flex flex-wrap items-center gap-md">
                Please follow below instruction for further steps
              </span>
              <CodeView
                preClassName="!overflow-none text-wrap break-words"
                copy
                data={data || ''}
              />
            </div>
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
          clusterName={item.displayName || ''}
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
            prefix={<ArrowClockwise size={16} />}
          />
        </div>
      )}
    </div>
  );
};

const ExtraButton = ({
  onDelete,
  onEdit,
  item,
}: {
  onDelete: () => void;
  onEdit: () => void;
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
}

const ClusterDnsView = ({ service }: { service: string }) => {
  return (
    <CopyContentToClipboard
      toolTip
      content={service}
      toastMessage="Cluster dns copied successfully."
    />
  );
};

const GridView = ({ items = [], onEdit, onDelete }: IResource) => {
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
                        item={item}
                      />
                    }
                    // action={
                    //   // <ExtraButton status={status.status} cluster={item} />
                    //   <span />
                    // }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, id + name + provider),
                render: () => (
                  <div className="flex flex-col gap-md">
                    {/* <ListItem data={path} /> */}
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
          {
            render: () => '',
            name: 'logs',
            className: 'min-w-[150px] flex-1 flex items-center justify-center',
          },
          {
            render: () => 'Provider',
            name: 'provider',
            className: 'flex w-[100px]',
          },
          {
            render: () => 'Dns',
            name: 'dns',
            className: 'flex w-[180px]',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px]',
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
          const { name, id, updateInfo, provider } = parseItem(i);

          const isLatest = dayjs(i.updateTime).isAfter(
            dayjs().subtract(3, 'hour')
          );

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
              logs: {
                render: () =>
                  i.type === 'normal'
                    ? isLatest && (
                        <Button
                          size="sm"
                          variant="basic"
                          content={open === i.id ? 'Hide Logs' : 'Show Logs'}
                          onClick={(e) => {
                            e.preventDefault();

                            setOpen((s) => {
                              if (s === i.id) {
                                return '';
                              }
                              return i.id;
                            });
                          }}
                        />
                      )
                    : i.type === 'byok' && <ByokButton item={i} />,
              },
              provider: { render: () => <ListItem data={provider} /> },
              dns: {
                render: () => (
                  <div className="flex w-fit truncate">
                    <ClusterDnsView service={`${parseName(i)}.local`} />
                  </div>
                ),
              },
              status: {
                render: () => <SyncStatusV2 item={i} />,
              },
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
                    item={i}
                  />
                ),
              },
            },
            ...(i.type === 'normal'
              ? { to: `/${account}/infra/${id}/overview` }
              : {}),
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

const ClusterResourcesV2 = ({
  items,
  byokItems,
}: {
  items: Omit<BaseType, 'type'>[];
  byokItems: Omit<ByokBaseType, 'type'>[];
}) => {
  const { account } = useOutletContext<IAccountContext>();
  const normalItems = useMapper(items, (i) => {
    return { ...i, type: 'normal' as BaseType['type'] };
  });

  const bItems = useMapper(byokItems, (i) => {
    return { ...i, type: 'byok' as ByokBaseType['type'] };
  });

  const finalItems = useAppend(normalItems, bItems);

  useWatchReload(
    finalItems.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(i)}`;
    })
  );

  const [showDeleteDialog, setShowDeleteDialog] =
    useState<CombinedBaseType | null>(null);
  const [showHandleByokCluster, setShowHandleByokCluster] =
    useState<CombinedBaseType | null>(null);

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items: finalItems,
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
          data: showHandleByokCluster! as ByokBaseType,
        }}
      />
    </>
  );
};

export default ClusterResourcesV2;
