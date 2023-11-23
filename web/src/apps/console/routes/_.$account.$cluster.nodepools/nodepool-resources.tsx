import { CodeSimple, PencilLine, Trash } from '@jengaicons/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
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
import { parseStatus } from '~/console/utils/commons';
import Popup from '~/components/molecule/popup';
import { HighlightJsLogs } from 'react-highlightjs-logs';
import { yamlDump } from '~/console/components/diff-viewer';
import DeleteDialog from '~/console/components/delete-dialog';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { Link } from '@remix-run/react';
import HandleNodePool from './handle-nodepool';

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

const ExtraButton = ({
  onDelete,
  onEdit,
  onShowResourceYaml,
}: {
  onDelete: () => void;
  onEdit: () => void;
  onShowResourceYaml: () => void;
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
          key: '2',
          label: 'Resource Yaml',
          icon: <CodeSimple size={16} />,
          type: 'item',
          onClick: onShowResourceYaml,
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
  onShowResourceYaml: (item: BaseType) => void;
}

const GridView = ({
  items,
  onDelete,
  onEdit,
  onShowResourceYaml,
}: IResource) => {
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
                  <ListTitleWithSubtitleAvatar
                    avatar={<ConsoleAvatar name={id} />}
                    title={name}
                    subtitle={id}
                    action={
                      <ExtraButton
                        onDelete={() => onDelete(item)}
                        onEdit={() => onEdit(item)}
                        onShowResourceYaml={() => onShowResourceYaml(item)}
                      />
                    }
                  />
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

const ListView = ({
  items,
  onDelete,
  onEdit,
  onShowResourceYaml,
}: IResource) => {
  return (
    <List.Root linkComponent={Link}>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            to={`../np/${id}`}
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'status'),
                className: 'w-[180px]',
                render: () => parseStatus(item).component,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[180px]',
                render: () => (
                  <ListItemWithSubtitle
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => (
                  <ExtraButton
                    onDelete={() => onDelete(item)}
                    onEdit={() => onEdit(item)}
                    onShowResourceYaml={() => onShowResourceYaml(item)}
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
    onShowResourceYaml: (item) => {
      setShowResourceYaml(item);
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
