import { PencilLine, Trash, Cpu } from '@jengaicons/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
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
import { Link } from '@remix-run/react';
import { IStatus, listRender } from '~/console/components/commons';
import AnimateHide from '~/components/atoms/animate-hide';
import HandleNodePool from './handle-nodepool';
import { findNodePlanWithCategory } from './nodepool-utils';

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
  status: _,
}: {
  onDelete: () => void;
  onEdit: () => void;
  status: IStatus;
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

const ListDetail = (
  props: Omit<IResource, 'items'> & { open: boolean; item: BaseType }
) => {
  const { item, open, onDelete, onEdit } = props;
  const { name, id } = parseItem(item);
  const { minCount, maxCount, cloudProvider, aws } = item.spec;
  const keyPrefix = `${RESOURCE_NAME}-${id}`;
  const lR = listRender({ keyPrefix, resource: item });

  const parseSize = () => {
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

  const parseProviderInfo = () => {
    const iconSize = 14;
    switch (cloudProvider) {
      case 'aws':
        const nodePlan = findNodePlanWithCategory(
          aws?.ec2Pool?.instanceType || ''
        );
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
      case 'do':
      case 'gcp':
      default:
        return null;
    }
  };

  const statusRender = lR.statusRender({ className: 'w-[180px]', type:'nodepool' });

  return (
    <div className="w-full flex flex-col">
      <div className="flex flex-row items-center">
        <div className="w-[220px] min-w-[220px] border-border-disabled border-r mr-xl">
          <ListTitle
            title={name}
            subtitle={id}
            avatar={<NodePoolAvatar title={id} />}
          />
        </div>
        <div className="flex flex-col gap-sm w-[150px] min-w-[150px] border-border-disabled border-r mr-xl pr-xl truncate">
          <span className="bodySm text-text-soft pulsable">Size</span>
          <span className="bodyMd-medium pulsable">{parseSize()}</span>
        </div>
        <div className="pr-7xl w-[200px] min-w-[200px]">
          <ListItem data={parseProviderInfo()} />
        </div>
        <div className="flex-grow flex items-center justify-center">
          {statusRender.render()}
        </div>
        {/* <div className="flex flex-row gap-2xl items-center pl-3xl pr-xl mr-3xl border-border-disabled border-r w-[160px] min-w-[160px]">
          <div className="flex flex-col gap-sm">
            <span className="bodySm text-text-soft pulsable">Expected</span>
            <span className="bodyMd-medium pulsable">
              {targetCount} node
              {targetCount > 1 && 's'}
            </span>
          </div>
          <div className="flex flex-col gap-sm">
            <span className="bodySm text-text-soft pulsable">Current</span>
            <span className="bodyMd-medium pulsable">
              {targetCount} node
              {targetCount > 1 && 's'}
            </span>
          </div>
        </div> */}
        <div className="pr-3xl w-[180px] min-w-[180px]">
          {lR.authorRender({ className: '' }).render()}
        </div>
        <ExtraButton
          onDelete={() => onDelete(item)}
          onEdit={() => onEdit(item)}
          status={statusRender.status}
        />
      </div>

      {/* <AnimateHide show={open}>hello</AnimateHide> */}
    </div>
  );
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
                        status="none"
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
  const [open, setOpen] = useState<string | null>(null);
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        // const lR = listRender({ keyPrefix, resource: item });
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            onClick={() => setOpen((prev) => (prev === id ? null : id))}
            // columns={[
            //   {
            //     key: generateKey(keyPrefix, name + id),
            //     className: 'flex-1',
            //     render: () => (
            //       <ListTitle
            //         title={name}
            //         subtitle={id}
            //         avatar={<NodePoolAvatar title={name} />}
            //       />
            //     ),
            //   },
            //   lR.statusRender({ className: 'w-[180px]' }),
            //   lR.authorRender({ className: 'w-[180px]' }),
            //   {
            //     key: generateKey(keyPrefix, 'action'),
            //     render: () => (
            //       <ExtraButton
            //         onDelete={() => onDelete(item)}
            //         onEdit={() => onEdit(item)}
            //         onShowResourceYaml={() => onShowResourceYaml(item)}
            //       />
            //     ),
            //   },
            // ]}
            columns={[
              {
                className: 'w-full',
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListDetail
                    item={item}
                    open={open === id}
                    onDelete={() => onDelete(item)}
                    onEdit={() => onEdit(item)}
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

const NodepoolResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showResourceYaml, setShowResourceYaml] = useState<BaseType | null>(
    null
  );
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showHandleNodepool, setShowHandleNodepool] = useState<BaseType | null>(
    null
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
        resourceName={showDeleteDialog?.displayName}
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

export default NodepoolResources;
