import { PencilLine, Trash, Cpu } from '~/console/components/icons';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { INodepools } from '~/console/server/gql/queries/nodepool-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useState } from 'react';
import Popup from '~/components/molecule/popup';
import { HighlightJsLogs } from 'react-highlightjs-logs';
import { yamlDump } from '~/console/components/diff-viewer';
import DeleteDialog from '~/console/components/delete-dialog';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { Link, useOutletContext, useParams } from '@remix-run/react';
import AnimateHide from '~/components/atoms/animate-hide';
import { Button } from '~/components/atoms/button';
import { dayjs } from '~/components/molecule/dayjs';
import LogComp from '~/root/lib/client/components/logger';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { IClusterContext } from '~/console/routes/_main+/$account+/infra+/$cluster+/_layout';
import LogAction from '~/console/page-components/log-action';
import { useDataState } from '~/console/page-components/common-state';
import ListV2 from '~/console/components/listV2';
import { SyncStatusV2 } from '~/console/components/sync-status';
import HandleNodePool from './handle-nodepool';
import {
  findNodePlanWithCategory,
  findNodePlanWithSpec,
} from './nodepool-utils';
import { IAccountContext } from '../../../_layout';

const RESOURCE_NAME = 'nodepool';
type BaseType = ExtractNodeType<INodepools>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ExtraButton = ({
  onDelete,
  onEdit,
}: //   status: _,
{
  onDelete: () => void;
  onEdit: () => void;
  //   status: IStatus;
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

const NodePoolAvatar = ({ title }: { title: string }) => {
  return (
    <div className="relative w-[28px] aspect-square">
      <div className="z-[5] absolute top-[100%] left-[50%] transform -translate-x-1/2 -translate-y-full">
        <ConsoleAvatar name={title} size="xs" />
      </div>
      <div className="absolute top-0 -left-sm">
        <ConsoleAvatar name="" color="#000" size="xs" />
      </div>
      <div className="absolute top-0 -right-sm">
        <ConsoleAvatar name="" color="#000" size="xs" />
      </div>
    </div>
  );
};

const parseSize = ({ minCount, maxCount }: BaseType['spec']) => {
  if (minCount === maxCount) {
    return (
      <div className="truncate">
        {minCount} node{minCount > 1 && 's'}
      </div>
    );
  }
  return (
    <div className="truncate">
      {minCount} - {maxCount} nodes
    </div>
  );
};

const parseProvider = ({ cloudProvider, aws, gcp }: BaseType['spec']) => {
  const iconSize = 14;
  switch (cloudProvider) {
    case 'aws':
      let nodePlan = findNodePlanWithCategory(aws?.ec2Pool?.instanceType || '');

      if (aws?.poolType === 'spot') {
        nodePlan = findNodePlanWithSpec({
          spot: true,
          spec: {
            cpu: aws.spotPool?.cpuNode?.vcpu.min,
            memory: aws.spotPool?.cpuNode?.memoryPerVcpu?.min,
          },
        });
        <div className="flex flex-col gap-sm">
          <div className="bodySm text-text-soft pulsable">
            {nodePlan?.category} - {nodePlan?.labelDetail.size}
          </div>
          <div className="flex flex-row gap-lg bodyMd-medium pulsable">
            <span className="flex flex-row gap-md items-center">
              <Cpu size={iconSize} /> <span>{nodePlan?.labelDetail.cpu}</span>
            </span>
            <span className="flex flex-row gap-md items-center">
              <Cpu size={iconSize} />{' '}
              <span>{nodePlan?.labelDetail.memory}</span>
            </span>
          </div>
        </div>;
      }

      return (
        <div className="flex flex-col gap-sm">
          <div className="bodySm text-text-soft pulsable">
            {nodePlan?.category} - {nodePlan?.labelDetail.size}
          </div>
          <div className="flex flex-row gap-lg bodyMd-medium pulsable">
            <span className="flex flex-row gap-md items-center">
              <Cpu size={iconSize} /> <span>{nodePlan?.labelDetail.cpu}</span>
            </span>
            <span className="flex flex-row gap-md items-center">
              <Cpu size={iconSize} />{' '}
              <span>{nodePlan?.labelDetail.memory}</span>
            </span>
          </div>
        </div>
      );
    case 'azure':
    case 'gcp':
      return (
        <div className="flex flex-col gap-sm w-[150px] min-w-[150px] truncate">
          <span className="bodySm text-text-soft pulsable">Machine type</span>
          <span className="bodyMd-medium pulsable">{gcp?.machineType}</span>
        </div>
      );
    default:
      return null;
  }
};

const ShowCodeInModal = ({
  text,
  visible,
  setVisible,
}: {
  text: string;
  visible: boolean;
  setVisible: (v: boolean) => void;
}) => {
  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      {/* <Popup.Header>Resource Yaml</Popup.Header> */}
      <Popup.Content className="!p-0">
        <HighlightJsLogs
          width="100%"
          height="40rem"
          title="Yaml Code"
          dark
          selectableLines
          text={text}
          language="yaml"
        />
      </Popup.Content>
    </Popup.Root>
  );
};

const GridView = ({ items, onDelete, onEdit }: IResource) => {
  return (
    <Grid.Root linkComponent={Link} className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            to={`../np/${id}`}
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitle
                    avatar={<ConsoleAvatar name={id} />}
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton
                        onDelete={() => onDelete(item)}
                        onEdit={() => onEdit(item)}
                      />
                    }
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

const ListView = ({ items, onDelete, onEdit }: IResource) => {
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
            className: 'min-w-[180px] flex-1 flex items-center justify-center',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'flex-1 min-w-[30px] flex items-center justify-center',
          },
          {
            render: () => 'Size',
            name: 'size',
            className: 'w-[150px]',
          },
          {
            render: () => 'Provider Info',
            name: 'provider',
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
          console.log('updateInfo', parseItem(i));
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
                    avatar={<NodePoolAvatar title={id} />}
                  />
                ),
              },
              logs: {
                render: () =>
                  isLatest ? (
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
                  ) : null,
              },
              status: {
                render: () => <SyncStatusV2 item={i} />,
              },
              size: {
                render: () => <ListItem data={parseSize(i.spec)} />,
              },
              provider: {
                render: () => <ListItem data={parseProvider(i.spec)} />,
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
                  />
                ),
              },
            },
            // to: `/${account}/deployment/${id}`,
            detail: (
              <AnimateHide show={open === i.id} className="w-full pt-4xl">
                <LogComp
                  {...{
                    dark: true,
                    width: '100%',
                    height: '40rem',
                    title: 'Logs',
                    hideLineNumber: !state.linesVisible,
                    hideTimestamp: !state.timestampVisible,
                    websocket: {
                      account: account || '',
                      cluster: i.clusterName,
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

const NodepoolResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const [showResourceYaml, setShowResourceYaml] = useState<BaseType | null>(
    null
  );
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleNodepool, setShowHandleNodepool] = useState<BaseType | null>(
    null
  );

  const { account } = useOutletContext<IAccountContext>();
  const { cluster } = useOutletContext<IClusterContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(
        cluster
      )}.nodepool:${parseName(i)}`;
    })
  );

  const reload = useReload();
  const api = useConsoleApi();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
    onEdit: (item) => {
      setShowHandleNodepool(item);
    },
  };
  return (
    <>
      <ListGridView
        gridView={<GridView {...props} />}
        listView={<ListView {...props} />}
      />
      <HandleNodePool
        {...{
          isUpdate: true,
          visible: !!showHandleNodepool,
          setVisible: () => {
            setShowHandleNodepool(null);
          },
          data: showHandleNodepool!,
        }}
      />

      <ShowCodeInModal
        visible={!!showResourceYaml}
        text={yamlDump(showResourceYaml!)}
        setVisible={() => setShowResourceYaml(null)}
      />

      <DeleteDialog
        resourceName={parseName(showDeleteDialog)}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteNodePool({
              clusterName: showDeleteDialog!.clusterName,
              poolName: parseName(showDeleteDialog),
            });

            if (errors) {
              throw errors[0];
            }
            reload();
            toast.success(`${titleCase(RESOURCE_NAME)} is added for deletion.`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};

export default NodepoolResourcesV2;
